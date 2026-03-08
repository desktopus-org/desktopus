'use strict';

const { app, BrowserWindow } = require('electron');
const path = require('path');

// In dev mode (`electron . <url>`), Electron inserts the app path at argv[1],
// so the user-supplied URL lands at argv[2].  In a packaged binary the app
// path is not injected, so the URL is at argv[1].  process.defaultApp is true
// only when launched via the Electron CLI (dev mode).
const url = (process.defaultApp ? process.argv[2] : process.argv[1]) || 'http://localhost:3000';

// Keys Electron would otherwise handle internally — suppress them so
// Selkies's renderer receives everything.
const SUPPRESSED = [
  (i) => i.control && i.key === 'r',
  (i) => i.control && i.shift && i.key === 'R',
  (i) => i.control && i.shift && i.key === 'I',
  (i) => i.control && i.shift && i.key === 'J',
  (i) => i.control && i.key === 'u',
  (i) => i.key === 'F12',
  (i) => i.key === 'F5',
  (i) => i.alt && i.key === 'ArrowLeft',
  (i) => i.alt && i.key === 'ArrowRight',
];

// navigator.keyboard.lock() is Chromium's API for capturing WM-level shortcuts.
// On X11 it calls XGrabKeyboard; on Wayland it uses the keyboard-shortcuts-inhibit
// protocol. Both intercept Alt+Tab and other compositor/WM grabs before they
// reach the host desktop.
//
// IMPORTANT: keyboard.lock() requires document.fullscreenElement to be non-null
// (renderer-side fullscreen). Electron kiosk mode only sets OS-level fullscreen,
// so we must call document.requestFullscreen() first, simulating a user gesture
// via executeJavaScript(code, true) to bypass the gesture requirement.
//
// keyboard.lock() itself is called from the enter-html-full-screen event handler,
// which fires after the OS has fully committed fullscreen and focus has settled —
// avoiding the timing race that occurs when calling lock() immediately after
// requestFullscreen() resolves.
const LOCK_JS = `
  (async () => {
    try {
      if (!document.fullscreenElement) {
        await document.documentElement.requestFullscreen();
      }
    } catch (_) {}
  })();
`;
const UNLOCK_JS = `
  if (navigator.keyboard && navigator.keyboard.unlock) {
    navigator.keyboard.unlock();
  }
  if (document.fullscreenElement) {
    document.exitFullscreen().catch(() => {});
  }
`;
const KEYBOARD_LOCK_JS = `
  if (navigator.keyboard && navigator.keyboard.lock) {
    navigator.keyboard.lock().catch(() => {});
  }
`;

function createWindow() {
  // Tracks whether renderer-side fullscreen+keyboard-lock is active.
  let captureActive = false;

  const win = new BrowserWindow({
    width: 1920,
    height: 1080,
    kiosk: false,
    autoHideMenuBar: true,
    title: 'desktopus-viewer',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      // Disable web security only for localhost targets — Selkies needs it.
      webSecurity: !url.startsWith('http://localhost'),
    },
  });

  win.loadURL(url);

  // Called after the OS has fully committed fullscreen and focus has settled.
  // This is the reliable moment to apply keyboard.lock().
  win.webContents.on('enter-html-full-screen', () => {
    captureActive = true;
    win.webContents.executeJavaScript(KEYBOARD_LOCK_JS, true).catch(() => {});
  });

  win.webContents.on('leave-html-full-screen', () => {
    captureActive = false;
    win.webContents.executeJavaScript(
      'if (navigator.keyboard && navigator.keyboard.unlock) { navigator.keyboard.unlock(); }',
      true
    ).catch(() => {});
  });

  // Re-lock on focus in case the OS released the grab (e.g. after a VT switch),
  // but only if the user has already entered capture mode.
  win.on('focus', () => {
    if (captureActive) {
      win.webContents.executeJavaScript(KEYBOARD_LOCK_JS, true).catch(() => {});
    }
  });

  win.webContents.on('before-input-event', (event, input) => {
    if (input.type !== 'keyDown') return;
    if (SUPPRESSED.some((fn) => fn(input))) {
      event.preventDefault();
    }
  });
}

app.whenReady().then(createWindow);

// Prevent navigation away from the target — the viewer is single-purpose.
app.on('web-contents-created', (_event, contents) => {
  contents.on('will-navigate', (event, navigationUrl) => {
    if (navigationUrl !== url) {
      event.preventDefault();
    }
  });
});
