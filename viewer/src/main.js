'use strict';

const { app, BrowserWindow } = require('electron');
const path = require('path');

const url = process.argv[2] || 'http://localhost:3000';

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
const LOCK_JS = `
  (async () => {
    try {
      if (!document.fullscreenElement) {
        await document.documentElement.requestFullscreen();
      }
      if (navigator.keyboard && navigator.keyboard.lock) {
        await navigator.keyboard.lock();
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

function lockKeyboard(win) {
  // userGesture=true is required so requestFullscreen() doesn't throw.
  win.webContents.executeJavaScript(LOCK_JS, true).catch(() => {});
}

function unlockKeyboard(win) {
  win.webContents.executeJavaScript(UNLOCK_JS, true).catch(() => {});
}

function createWindow() {
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

  // Re-lock on focus in case the OS released the grab (e.g. after a VT switch),
  // but only if the user has already entered capture mode.
  win.on('focus', () => {
    if (win.isKiosk()) lockKeyboard(win);
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
