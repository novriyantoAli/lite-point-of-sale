import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

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
