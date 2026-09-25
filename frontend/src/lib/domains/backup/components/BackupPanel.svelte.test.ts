import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import QueryClientHarness from '$lib/testing/QueryClientHarness.svelte';
import BackupPanel from './BackupPanel.svelte';

const { list, create } = vi.hoisted(() => ({
	list: vi.fn(),
	create: vi.fn()
}));

// The api layer is the seam: components never see axios (ADR-0007).
vi.mock('../api/backup.api', () => ({
	backupApi: { list, create }
}));

const backup = {
	name: 'pos-20260115-120000-000000000.db',
	size: 8192,
	created_at: '2026-01-15T12:00:00Z'
};

function renderPanel() {
	return render(BackupPanel, {}, { wrapper: QueryClientHarness });
}

describe('BackupPanel', () => {
	beforeEach(() => {
		list.mockReset();
		create.mockReset();
	});

	it('lists the snapshots the store has kept', async () => {
		list.mockResolvedValue([backup]);

		renderPanel();

		expect(await screen.findByText(backup.name)).toBeInTheDocument();
	});

	it('takes a manual backup and reports the file it created', async () => {
		list.mockResolvedValue([]);
		create.mockResolvedValue(backup);
		const user = userEvent.setup();

		renderPanel();

		await user.click(screen.getByRole('button', { name: 'Backup sekarang' }));

		expect(create).toHaveBeenCalled();
		expect(await screen.findByRole('status')).toHaveTextContent(`Backup dibuat: ${backup.name}`);
	});

	it('shows the message the API sent when the export is refused', async () => {
		list.mockResolvedValue([]);
		create.mockRejectedValue({ message: 'Terjadi kesalahan pada server.', status: 500 });
		const user = userEvent.setup();

		renderPanel();

		await user.click(screen.getByRole('button', { name: 'Backup sekarang' }));

		expect(await screen.findByRole('alert')).toHaveTextContent('Terjadi kesalahan pada server.');
	});

	it('shows the normalized error and a retry when the list fails', async () => {
		list.mockRejectedValue({
			message: 'Anda tidak berhak melakukan tindakan ini.',
			status: 403
		});
		const user = userEvent.setup();

		renderPanel();

		expect(
			await screen.findByText('Anda tidak berhak melakukan tindakan ini.')
		).toBeInTheDocument();

		await user.click(screen.getByRole('button', { name: 'Coba lagi' }));

		expect(list).toHaveBeenCalledTimes(2);
	});

	it('says so when there are no backups yet, instead of showing nothing', async () => {
		list.mockResolvedValue([]);

		renderPanel();

		expect(await screen.findByText(/Belum ada backup/)).toBeInTheDocument();
	});
});
