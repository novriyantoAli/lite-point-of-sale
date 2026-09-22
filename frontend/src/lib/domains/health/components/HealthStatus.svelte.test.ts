import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import HealthStatus from './HealthStatus.svelte';

const { check } = vi.hoisted(() => ({ check: vi.fn() }));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/health.api', () => ({ healthApi: { check } }));

describe('HealthStatus', () => {
	beforeEach(() => {
		check.mockReset();
	});

	it('shows OK when the API reports a healthy service', async () => {
		check.mockResolvedValue({ status: 'ok', database: 'ok' });

		render(HealthStatus, {}, { wrapper: QueryClientHarness });

		expect(await screen.findByText('OK')).toBeInTheDocument();
		expect(screen.getByText(/Basis data: ok/)).toBeInTheDocument();
	});

	it('shows DEGRADED when the database is unavailable', async () => {
		check.mockResolvedValue({ status: 'degraded', database: 'unavailable' });

		render(HealthStatus, {}, { wrapper: QueryClientHarness });

		expect(await screen.findByText('DEGRADED')).toBeInTheDocument();
	});

	it('tells the user when the server cannot be reached', async () => {
		check.mockRejectedValue(new Error('backend unreachable'));

		render(HealthStatus, {}, { wrapper: QueryClientHarness });

		expect(await screen.findByText('Tidak dapat menghubungi server.')).toBeInTheDocument();
	});
});
