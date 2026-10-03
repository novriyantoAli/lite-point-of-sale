<script lang="ts">
	import {
		AKSI_MODUL,
		CATATAN,
		CATATAN_ANGKA,
		COMMIT_BARIS,
		DAFTAR,
		ERROR,
		FIELD,
		INPUT,
		JUDUL,
		LABEL,
		MODUL,
		MODUL_BARIS,
		MODUL_JUDUL,
		MODUL_KEPALA,
		MODUL_TEKS,
		PAPAN,
		STRIP,
		TAG_CORET
	} from '$lib/components/shared/mosaik';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import { cn, collectFieldErrors } from '$lib/utils';
	import RoleBadge, { PERAN_LABEL } from './RoleBadge.svelte';
	import {
		createPenggunaListQuery,
		createPenggunaMutation,
		createSetPenggunaActiveMutation
	} from '../queries/auth.queries';
	import { CreatePenggunaInputSchema, RoleSchema, type Role } from '../schemas/auth.schema';

	/**
	 * `currentUserId` is the Pengguna using the page. It comes from the layout's
	 * server load, so the row of the Admin themselves can be shown as untouchable
	 * instead of offering a button the API would refuse.
	 */
	let { currentUserId }: { currentUserId: number } = $props();

	const pengguna = createPenggunaListQuery();
	const create = createPenggunaMutation();
	const setActive = createSetPenggunaActiveMutation();

	let username = $state('');
	let password = $state('');
	let role = $state<Role>('kasir');
	let fieldErrors = $state<Partial<Record<'username' | 'password', string>>>({});
	let notice = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		fieldErrors = {};
		notice = '';

		// The same schema the api layer parses with, so a rule is written once.
		const parsed = CreatePenggunaInputSchema.safeParse({ username, password, role });
		if (!parsed.success) {
			fieldErrors = collectFieldErrors(parsed.error, ['username', 'password'] as const);
			return;
		}

		try {
			const created = await create.mutateAsync(parsed.data);
			notice = `Pengguna ${created.username} ditambahkan.`;
			username = '';
			/*
			 * The password is cleared with the rest of the form: it was typed here to
			 * be sent once, and nothing on this screen ever holds it again — the list
			 * below shows username, Peran, and whether the account is Aktif, and the
			 * API answers a Pengguna without a password either (auth.schema.ts).
			 */
			password = '';
			role = 'kasir';
		} catch {
			// create.error carries the normalized message, rendered below.
		}
	}

	/**
	 * The two Peran, in the order the picker offers them. It is the schema's own list
	 * rather than a copy of it — `RoleSchema` mirrors the Go enum, so the order comes
	 * from the contract — and the words come from `RoleBadge`, the one place a Peran
	 * is written for a person.
	 */
	const ROLES = RoleSchema.options;

	/**
	 * Jarak 8px sebelum modul Daftar bukan kelalaian: prototipe layar ini sendiri
	 * yang memisahkan kedua section dengan `margin-top: 8px`.
	 */
	const MODUL_DAFTAR = cn(MODUL, 'mt-2');
	/** Fields & Inputs: field 26px, label 12px/600 di atasnya dengan jarak 3px. */
	const FORMGRID = 'grid grid-cols-1 gap-x-3 gap-y-[10px] min-[901px]:grid-cols-3';
	const FIELD_SELECT = cn(
		INPUT,
		'w-full data-[size=default]:h-[26px] [&_svg:not([class*=size-])]:size-3'
	);
	/**
	 * Baris pesan — galat maupun hasil — di dalam modul ini, dan warnanya tinta,
	 * bukan merah utilitas. Merah hanya dua pekerjaan, tab dan harga (DESIGN.md,
	 * Secondary & Three-Percent Rule); yang memakai merah hanyalah garis dan
	 * outline field yang tidak valid, dan keputusan itu sudah berlaku seragam di
	 * layar terporting (#46).
	 */
	const PESAN = ERROR;
	/**
	 * Satu baris pesan yang berdiri di luar modul, tepat di bawahnya: sisi dan
	 * bawahnya digambar, atasnya tidak — garis yang memisahkannya dari modul di
	 * atasnya sudah digambar modul itu, dan satu garis dipakai bersama, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const BARIS_GALAT = 'border border-t-0 border-border bg-card px-2 py-1.5';
	/** Baris aksi di bawah kisi field: `gap: 4px` dan `margin-top: 10px` prototipe. */
	const AKSI = 'mt-[10px] flex flex-wrap items-center gap-1';
	const LIROW = 'flex flex-wrap items-center justify-between gap-x-3 gap-y-1.5 px-2 py-1.5';
	const LIROW_UTAMA = 'min-w-0';
	/**
	 * Nama Pengguna dan tanda-tandanya dalam satu baris. Jaraknya datang dari `gap`,
	 * bukan dari spasi di dalam markah, supaya tag Peran yang 15px tidak menempel ke
	 * namanya saat baris itu membungkus.
	 *
	 * Username tetap tinggal di simpulnya sendiri: suite e2e membacanya sebagai nama
	 * yang persis (`penggunaRow`), dan yang boleh berubah di port ini bentuknya,
	 * bukan kalimatnya (solution doc §9).
	 */
	const LIROW_NAMA = 'flex flex-wrap items-center gap-x-2 gap-y-1 text-[13px] font-medium';
	const LIROW_AKSI = 'flex items-center gap-1';
</script>

<div class={PAPAN}>
	<!--
		Strip judul: judul 24px/700 dengan satu kalimat 12px di sebelahnya
		(DESIGN.md, Typography). Layar ini tidak punya aksi utama di strip — yang
		menambahkan Pengguna adalah formulir di bawahnya, yang memang pekerjaan
		layar ini.
	-->
	<div class={STRIP}>
		<h1 class={JUDUL}>Pengguna</h1>
		<p class={CATATAN}>
			Kasir dan Admin yang boleh memakai kasir ini. Akun yang Nonaktif tetap tersimpan riwayatnya,
			tetapi tidak bisa login lagi.
		</p>
	</div>

	<section class={MODUL} aria-labelledby="tambah-pengguna-judul">
		<div class={MODUL_KEPALA}>
			<h2 id="tambah-pengguna-judul" class={MODUL_JUDUL}>Tambah Pengguna</h2>
			<span class={CATATAN}>Admin membuat akun Kasir</span>
		</div>

		<form class={MODUL_BARIS} onsubmit={submit} novalidate>
			<div class={FORMGRID}>
				<div class={FIELD}>
					<Label for="pengguna-username" class={LABEL}>Username</Label>
					<Input
						id="pengguna-username"
						name="username"
						autocomplete="off"
						class={INPUT}
						bind:value={username}
						aria-invalid={fieldErrors.username ? true : undefined}
					/>
					{#if fieldErrors.username}
						<p class={PESAN}>{fieldErrors.username}</p>
					{/if}
				</div>

				<div class={FIELD}>
					<Label for="pengguna-password" class={LABEL}>Password</Label>
					<Input
						id="pengguna-password"
						name="password"
						type="password"
						autocomplete="new-password"
						class={INPUT}
						bind:value={password}
						aria-invalid={fieldErrors.password ? true : undefined}
					/>
					{#if fieldErrors.password}
						<p class={PESAN}>{fieldErrors.password}</p>
					{/if}
				</div>

				<div class={FIELD}>
					<Label for="pengguna-role" class={LABEL}>Peran</Label>
					<Select.Root type="single" bind:value={role}>
						<Select.Trigger id="pengguna-role" class={FIELD_SELECT}>
							<!--
								The label is written here rather than left to `Select.Value`: the
								label registry is filled by items as they mount, and the content
								mounts lazily, so before the dropdown is ever opened the trigger
								would show the raw value — `kasir`, a word this screen never says
								(#33).
							-->
							<span data-slot="select-value">{PERAN_LABEL[role]}</span>
						</Select.Trigger>
						<Select.Content
							class="rounded-none border border-border shadow-none ring-0 ring-transparent"
						>
							{#each ROLES as value (value)}
								<Select.Item {value} label={PERAN_LABEL[value]}>{PERAN_LABEL[value]}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
			</div>

			{#if create.error}
				<p class={cn(PESAN, 'mt-[10px]')} role="alert">{create.error.message}</p>
			{/if}

			<div class={AKSI}>
				<Button class={COMMIT_BARIS} type="submit" disabled={create.isPending}>
					{create.isPending ? 'Menyimpan…' : 'Tambah'}
				</Button>
				<!--
					Pemberitahuan hasil berdiri di sebelah tombolnya sendiri: ia menjawab
					"Tambah" itu, bukan daftar di bawahnya — dan daftar itu tetap tidak
					menerima baris kedua yang menyebut Pengguna yang sama (bandingkan #34).
				-->
				{#if notice}
					<p class={PESAN} role="status">{notice}</p>
				{/if}
			</div>
		</form>
	</section>

	<section
		class={MODUL_DAFTAR}
		aria-labelledby="daftar-pengguna-judul"
		aria-busy={pengguna.isPending}
	>
		<div class={MODUL_KEPALA}>
			<h2 id="daftar-pengguna-judul" class={MODUL_JUDUL}>Daftar Pengguna</h2>
			<span class={CATATAN_ANGKA}>
				{pengguna.isPending
					? 'memuat…'
					: pengguna.error
						? '—'
						: `${pengguna.data?.length ?? 0} akun`}
			</span>
		</div>

		{#if pengguna.isPending}
			<!-- Keadaan memuat menyebut dirinya dengan kata, bukan layar kosong. -->
			<p class={MODUL_TEKS}>Memuat daftar Pengguna…</p>
		{:else if pengguna.error}
			<div class={MODUL_BARIS}>
				<p class={PESAN} role="alert">{pengguna.error.message}</p>
				<Button
					variant="ghost"
					class={cn(AKSI_MODUL, 'mt-1.5')}
					onclick={() => void pengguna.refetch()}>Coba lagi</Button
				>
			</div>
		{:else if pengguna.data?.length === 0}
			<p class={MODUL_TEKS}>Belum ada Pengguna lain. Tambahkan Kasir lewat formulir di atas.</p>
		{:else}
			<ul class={DAFTAR}>
				{#each pengguna.data ?? [] as user (user.id)}
					<li class={LIROW}>
						<div class={LIROW_UTAMA}>
							<p class={LIROW_NAMA}>
								<span>{user.username}</span>
								<!-- Peran memakai tag yang sudah ada, bukan bentuk kedua. -->
								<RoleBadge role={user.role} />
								{#if !user.active}
									<span class={TAG_CORET}>Nonaktif</span>
								{/if}
								{#if user.id === currentUserId}
									<!--
										Kata `(Anda)` tinggal di simpulnya sendiri, terpisah dari
										pemisah `·`: satu tes komponen membacanya sebagai teks yang
										persis, dan kalimatnya tidak boleh berubah di port ini
										(solution doc §9).
									-->
									<span class={CATATAN}>· <span>(Anda)</span></span>
								{/if}
							</p>
						</div>

						<div class={LIROW_AKSI}>
							{#if user.id === currentUserId}
								<!--
									Larangannya berdiri sebagai kalimat, bukan sebagai tombol yang
									diam-diam tidak bekerja, dan bukan sebagai warna: Admin yang
									mencoba menonaktifkan akunnya sendiri harus tahu kenapa ia tidak
									bisa. Aturannya sendiri milik backend — layar ini tidak
									mengulangnya, ia hanya menuliskannya (issue #35).
								-->
								<span class={CATATAN}>Tidak bisa menonaktifkan diri sendiri</span>
							{:else}
								<Button
									variant="ghost"
									class={AKSI_MODUL}
									disabled={setActive.isPending}
									onclick={() => setActive.mutate({ id: user.id, active: !user.active })}
								>
									{user.active ? 'Nonaktifkan' : 'Aktifkan'}
								</Button>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<!--
		Galat Nonaktifkan berdiri di luar modul Daftar, sebagai barisnya sendiri:
		ia berlaku untuk satu aksi di dalam daftar, dan menaruhnya di dalam baris
		Pengguna yang digugurkan akan menghapus nama Pengguna yang sedang dibicarakan
		alih-alih menjelaskannya.
	-->
	{#if setActive.error}
		<p class={cn(BARIS_GALAT, PESAN)} role="alert">{setActive.error.message}</p>
	{/if}
</div>
