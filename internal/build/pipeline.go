package build

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/moby/moby/client"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
	"github.com/desktopus-org/desktopus/internal/progress"
)

//go:embed templates/*
var templatesFS embed.FS

// Options configures the build
type Options struct {
	Tag              string
	NoCache          bool
	AnsibleVerbosity int
}

// Pipeline orchestrates the full image build
type Pipeline struct {
	docker   *client.Client
	registry *module.Registry
}

// NewPipeline creates a new build pipeline
func NewPipeline(docker *client.Client, registry *module.Registry) *Pipeline {
	return &Pipeline{
		docker:   docker,
		registry: registry,
	}
}

// Build builds a Docker image from a desktop config
func (p *Pipeline) Build(ctx context.Context, cfg *config.DesktopConfig, configDir string, opts Options, output io.Writer) error {
	// 1. Resolve all modules
	modules, varsOverrides, err := p.resolveModules(cfg, configDir)
	if err != nil {
		return fmt.Errorf("resolving modules: %w", err)
	}

	// 2. Parse templates
	tmplFuncs := templateFuncs()
	dockerfileTmpl, err := template.New("Dockerfile.tmpl").Funcs(tmplFuncs).ParseFS(templatesFS, "templates/Dockerfile.tmpl")
	if err != nil {
		return fmt.Errorf("parsing Dockerfile template: %w", err)
	}
	playbookTmpl, err := template.New("playbook.yml.tmpl").ParseFS(templatesFS, "templates/playbook.yml.tmpl")
	if err != nil {
		return fmt.Errorf("parsing playbook template: %w", err)
	}
	ansibleCfgTmpl, err := template.New("ansible.cfg.tmpl").ParseFS(templatesFS, "templates/ansible.cfg.tmpl")
	if err != nil {
		return fmt.Errorf("parsing ansible.cfg template: %w", err)
	}

	// 3. Generate files
	dockerfile, err := generateDockerfile(dockerfileTmpl, cfg, modules, opts.AnsibleVerbosity)
	if err != nil {
		return err
	}
	playbook, err := generatePlaybook(playbookTmpl, modules, varsOverrides, cfg.Base.OS)
	if err != nil {
		return err
	}
	ansibleCfg, err := generateAnsibleCfg(ansibleCfgTmpl)
	if err != nil {
		return err
	}

	// 4. Assemble build context
	bctx := NewBuildContext()

	if err := bctx.AddFile("Dockerfile", dockerfile, 0644); err != nil {
		return fmt.Errorf("adding Dockerfile: %w", err)
	}
	if err := bctx.AddFile("playbook.yml", playbook, 0644); err != nil {
		return fmt.Errorf("adding playbook.yml: %w", err)
	}
	if err := bctx.AddFile("ansible.cfg", ansibleCfg, 0644); err != nil {
		return fmt.Errorf("adding ansible.cfg: %w", err)
	}

	// 5. Add module files to build context
	if err := p.addModuleFiles(bctx, modules); err != nil {
		return fmt.Errorf("adding module files: %w", err)
	}

	// 6. Add post-run scripts
	if err := addPostRunScripts(bctx, cfg); err != nil {
		return fmt.Errorf("adding post-run scripts: %w", err)
	}

	// 7. Add runtime files
	if err := addRuntimeFiles(bctx, cfg); err != nil {
		return fmt.Errorf("adding runtime files: %w", err)
	}

	reader, err := bctx.Reader()
	if err != nil {
		return fmt.Errorf("finalizing build context: %w", err)
	}

	// 8. Build image via Docker SDK
	imageTag := cfg.ImageTag()
	if opts.Tag != "" {
		imageTag = opts.Tag
	}

	resp, err := p.docker.ImageBuild(ctx, reader, client.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    opts.NoCache,
	})
	if err != nil {
		return fmt.Errorf("starting build: %w", err)
	}
	defer resp.Body.Close()

	// 9. Stream build output
	return streamBuildOutput(resp.Body, output)
}

func (p *Pipeline) resolveModules(cfg *config.DesktopConfig, configDir string) ([]*module.Module, []map[string]interface{}, error) {
	modules := make([]*module.Module, 0, len(cfg.Modules))
	varsOverrides := make([]map[string]interface{}, 0, len(cfg.Modules))

	for _, ref := range cfg.Modules {
		mod, err := p.registry.Resolve(ref.Name, configDir)
		if err != nil {
			return nil, nil, err
		}

		if !mod.IsCompatible(cfg.Base.OS, cfg.Base.Desktop) {
			return nil, nil, fmt.Errorf("module %q is not compatible with %s/%s", mod.Name, cfg.Base.OS, cfg.Base.Desktop)
		}

		modules = append(modules, mod)
		varsOverrides = append(varsOverrides, ref.Vars)
	}

	return modules, varsOverrides, nil
}

func (p *Pipeline) addModuleFiles(bctx *BuildContext, modules []*module.Module) error {
	builtinFS := p.registry.BuiltinFS()

	for _, mod := range modules {
		var srcFS fs.FS
		var root string

		if mod.Builtin {
			srcFS = builtinFS
			root = mod.Path
		} else {
			srcFS = dirFS(mod.Path)
			root = "."
		}

		if err := fs.WalkDir(srcFS, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			data, err := fs.ReadFile(srcFS, path)
			if err != nil {
				return err
			}

			// For built-in modules, path is already relative to builtinFS root
			// e.g., "chrome/tasks/main.yml"
			destPath := filepath.Join("modules", path)
			if !mod.Builtin {
				// For custom modules, path starts with "." so we remap
				relPath, _ := filepath.Rel(root, path)
				destPath = filepath.Join("modules", mod.Name, relPath)
			}

			return bctx.AddFile(destPath, data, 0644)
		}); err != nil {
			return fmt.Errorf("walking module %q: %w", mod.Name, err)
		}
	}

	return nil
}

func addPostRunScripts(bctx *BuildContext, cfg *config.DesktopConfig) error {
	for i, pr := range cfg.PostRun {
		runas := pr.RunAs
		if runas == "" {
			runas = "abc"
		}

		var script string
		if runas == "root" {
			script = fmt.Sprintf("#!/usr/bin/with-contenv bash\n# Desktopus post-run: %s\n\n%s\n", pr.Name, pr.Script)
		} else {
			script = fmt.Sprintf("#!/usr/bin/with-contenv bash\n# Desktopus post-run: %s (runas: %s)\n\nexec s6-setuidgid %s bash -c '\n%s\n'\n", pr.Name, runas, runas, pr.Script)
		}

		filename := fmt.Sprintf("postrun/%02d-desktopus-%s.sh", 50+i, pr.Name)
		if err := bctx.AddFile(filename, []byte(script), 0755); err != nil {
			return err
		}
	}
	return nil
}

func addRuntimeFiles(bctx *BuildContext, cfg *config.DesktopConfig) error {
	if len(cfg.Files) == 0 {
		return nil
	}

	// Add file templates
	for _, f := range cfg.Files {
		destPath := filepath.Join("runtime-files", f.Path)
		if err := bctx.AddFile(destPath, []byte(f.Content), 0644); err != nil {
			return err
		}
	}

	// Generate the provisioner script
	var script string
	script = "#!/usr/bin/with-contenv bash\n# Desktopus runtime file provisioner\nset -e\n\n"
	for _, f := range cfg.Files {
		mode := f.Mode
		if mode == "" {
			mode = "0644"
		}
		script += fmt.Sprintf("mkdir -p \"$(dirname '%s')\"\n", f.Path)
		script += fmt.Sprintf("envsubst < '/tmp/desktopus-runtime-files%s' > '%s'\n", f.Path, f.Path)
		script += fmt.Sprintf("chown abc:abc '%s'\n", f.Path)
		script += fmt.Sprintf("chmod %s '%s'\n\n", mode, f.Path)
	}

	return bctx.AddFile("99-desktopus-files.sh", []byte(script), 0755)
}

func streamBuildOutput(reader io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(reader)
	pr := progress.New(output)
	for {
		var msg struct {
			Stream   string `json:"stream"`
			Error    string `json:"error"`
			Status   string `json:"status"`
			Progress string `json:"progress"`
			ID       string `json:"id"`
		}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if msg.Error != "" {
			pr.Clear()
			return fmt.Errorf("build error: %s", msg.Error)
		}
		if msg.Stream != "" {
			pr.Flush()
			fmt.Fprint(output, msg.Stream)
			continue
		}
		if msg.Status != "" {
			if msg.ID != "" {
				pr.Update(msg.ID, msg.Status, msg.Progress)
			} else {
				pr.Print(msg.Status)
			}
		}
	}
}

// dirFS returns an os.DirFS for the given path
func dirFS(path string) fs.FS {
	return os.DirFS(path)
}
