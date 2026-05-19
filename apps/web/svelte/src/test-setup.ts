import '@testing-library/jest-dom/vitest';

// happy-dom doesn't implement HTMLVideoElement.play/load/pause — stub them
// so component tests that mount Player don't crash on missing methods.
Object.defineProperty(window.HTMLMediaElement.prototype, 'play', {
  writable: true,
  value: () => Promise.resolve(),
});
Object.defineProperty(window.HTMLMediaElement.prototype, 'load', {
  writable: true,
  value: () => undefined,
});
Object.defineProperty(window.HTMLMediaElement.prototype, 'pause', {
  writable: true,
  value: () => undefined,
});
