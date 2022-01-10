#!/bin/bash
# Build image
# Read all values from csv and create workspaces
WORKDIR="$(pwd)"
for workspace_info in workspaces/*
do
    # Getting workspace name and base image
    workspace_name=$(basename "${workspace_info##*/}" .info)
    distro=$(grep 'Base' "${workspace_info}" | cut -d'=' -f2 | cut -d':' -f1)
    distro_version=$(grep 'Base' "${workspace_info}" | cut -d'=' -f2 | cut -d':' -f2)
    echo "Building image '${workspace_name}' based on: '${distro}:${distro_version}'..."

    # Move base images to temporary folder
    TMP_DIR=$(mktemp -d -t "${workspace_name}"-XXXXXXXXXX)
    cp -r "src/${distro}/base/${distro_version}/." "${TMP_DIR}"
    mkdir -p "${TMP_DIR}/modules-installation"
    mkdir -p "${TMP_DIR}/modules-persistence"
    echo "${TMP_DIR}"

    # Fill template of run_workspace.sh and create_image.sh
    sed -i "s|__workspace_name__|${workspace_name}|g" "${TMP_DIR}/create_image.sh"
    sed -i "s|__distro_and_version__|${distro}-${distro_version}|g" "${TMP_DIR}/create_image.sh"
    sed -i "s|__workspace_name__|${workspace_name}|g" "${TMP_DIR}/run_workspace.sh"
    chmod +x "${TMP_DIR}/create_image.sh" "${TMP_DIR}/run_workspace.sh"


    # Persistent volumes for the workspace
    VOLUMES_COMMAND_LIST=""
    DOCKER_CP_INIT_LIST_COMMANDS=""
    MKDIR_MODULES_COMMANDS=""

    # Read all modules
    MODULES_LIST=$(grep 'Modules' "${workspace_info}" | cut -d'=' -f2)
    for module in ${MODULES_LIST//,/ }
    do
        module_name=$(echo "${module}" | cut -d':' -f1)
        
        # If ':' appears at least once, there is a version defined
        num_double_dots=$(echo "${module}" | tr -cd ':' | wc -c)
        if [ "${num_double_dots}" -eq 1 ]; then
            module_version=$(echo "${module}" | cut -d':' -f2)
            echo "    Module Name: ${module_name} - Version: ${module_version}"
        else 
            echo "    Module Name: ${module_name} - Version: No version specified"
        fi

        # Create temporal folder to process module
        cp -r "src/${distro}/modules/${module_name}" "${TMP_DIR}"
        module_temp_dir="${TMP_DIR}/${module_name}_tmp"
        mv "${TMP_DIR}/${module_name}" "${module_temp_dir}"

        cd "${module_temp_dir}" || exit 1
        # Create module directoty
        for module_file in *; do
            if [[ -f "${module_file}" ]]; then
                if [[ -n ${module_version} ]]; then
                    echo "        Replacing __version__ in ${module_file}"
                    sed -i "s|__version__|${module_version}|g" "${module_file}"
                fi
                # Move files to base image for installation
                if [[ "${module_file}" == "installation.sh" ]]; then
                    mv "${module_file}" "${module_name}_installation.sh"
                    mv "${module_name}_installation.sh" "${TMP_DIR}/modules-installation"
                fi
                if [[ "${module_file}" == "supervisor.conf" ]]; then
                    mv "${module_file}" "${module_name}.conf"
                    mv "${module_name}.conf" "${TMP_DIR}/supervisor"
                fi
                if [[ "${module_file}" == "file-desktop.desktop" ]]; then
                    mkdir -p "${TMP_DIR}/config-files/Desktop"
                    mv "${module_file}" "${module_name}.desktop"
                    mv "${module_name}.desktop" "${TMP_DIR}/config-files/Desktop"
                fi
                if [[ "${module_file}" == "module.info" ]]; then
                    # Check persistent volumes to be generated
                    persitent_volume_list=$(grep 'PersistentVolumes' "${module_file}" | cut -d'=' -f2) 
                    for persistent_volume in ${persitent_volume_list//,/ }
                    do
                        VOLUMES_COMMAND_LIST+="--volume \"\$(pwd)\"/modules-persistence/${module_name}${persistent_volume}:${persistent_volume} "
                        MKDIR_MODULES_COMMANDS+="mkdir -p \"modules-persistence/${module_name}${persistent_volume}\" \&\& "
                        DOCKER_CP_INIT_LIST_COMMANDS+="docker cp \"${workspace_name}:${persistent_volume}/.\" \"modules-persistence/${module_name}${persistent_volume}\" ; "
                    done
                fi
            fi
        done
        # Remove temporal module processing directory
        cd "${WORKDIR}" || exit 1
        rm -rf "${module_temp_dir}"
    done
    cd "${WORKDIR}" || exit 1
    # Replace modules persistence
    MKDIR_MODULES_COMMANDS+="true"
    sed -i "s|__modules_persistent_volumes__|${VOLUMES_COMMAND_LIST}|g" "${TMP_DIR}/run_workspace.sh"
    sed -i "s|__mkdir_modules_init__|${MKDIR_MODULES_COMMANDS}|g" "${TMP_DIR}/run_workspace.sh"
    sed -i "s|__docker_cp_init__|${DOCKER_CP_INIT_LIST_COMMANDS}|g" "${TMP_DIR}/run_workspace.sh"
    mkdir -p generated_workspaces_zips generated_workspaces
    cd "${TMP_DIR}" || exit 1
    zip -r "${WORKDIR}/generated_workspaces_zips/${workspace_name}.zip" ./*
    cd "${WORKDIR}" || exit 1
    cp -r "${TMP_DIR}/." "generated_workspaces/${workspace_name}"
    rm -rf "${TMP_DIR}"
done