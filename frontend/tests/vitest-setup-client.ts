import '@testing-library/jest-dom/vitest';
import { afterEach, vi } from 'vitest';

// jsdom does not implement matchMedia; Svelte 5 components (and the shadcn
// primitives) expect it to exist.
Object.defineProperty(window, 'matchMedia', {
	writable: true,
	enumerable: true,
	value: vi.fn().mockImplementation((query: string) => ({
		matches: false,
		media: query,
		onchange: null,
		addListener: vi.fn(),
		removeListener: vi.fn(),
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
		dispatchEvent: vi.fn()
	}))
});

// jsdom does not implement scrollIntoView either.
Element.prototype.scrollIntoView = vi.fn();

// Nor the pointer-capture API, which bits-ui's Select calls on pointerdown to
// keep a drag inside the trigger. Without these, opening a Select throws before
// an option can be picked and no component test can reach one.
Element.prototype.hasPointerCapture = vi.fn();
Element.prototype.setPointerCapture = vi.fn();
Element.prototype.releasePointerCapture = vi.fn();

// bits-ui's overlays (Dialog, Select) lock the page while they are open, and
// they hand the body back on a timer rather than on unmount. A test that starts
// inside that window inherits the lock from the test before it and cannot click
// anything — a leak in the harness, not in the component under test. Clearing
// the body's inline style puts every test back on the same footing.
afterEach(() => {
	document.body.removeAttribute('style');
});
