<script lang="ts">
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
	 * Papan mosaik (DESIGN.md, Layout): satu kolom selebar papan — layar Pengguna
	 * bukan papan tiga kolom milik Kasir. Padding luarnya milik rel
	 * (`(app)/+layout`), jadi di sini tidak ada padding lagi.
	 */
	const PAPAN = 'grid grid-cols-1 gap-0';
	/**
	 * Strip judul: satu-satunya tempat ukuran 24px muncul di layar ini. `-mb-px`
	 * plus `z-[2]` menariknya turun satu piksel, jadi garis tintanya yang menutup
	 * modul di bawahnya alih-alih bertumpuk dengan garis rambut (DESIGN.md,
	 * Shared-Hairline).
	 */
	const STRIP =
		'z-[2] col-span-full -mb-px flex flex-wrap items-baseline gap-x-3 gap-y-1 border border-border border-b-foreground bg-card px-2 py-2';
	const JUDUL = 'text-2xl leading-none font-bold tracking-[-0.015em]';
	/** Seluruh teks 12px; lantai huruf dunia ini (DESIGN.md, Legibility Floor). */
	const CATATAN = 'text-xs';
	/**
	 * Cacah akun di kepala modul Daftar. `tabular-nums` bukan hiasan: ia cacah yang
	 * berubah saat Pengguna ditambahkan, dan DESIGN.md menuntutnya untuk "setiap
	 * angka uang, jumlah, dan Stok" — prototipe menandai catatan kepala yang sama
	 * `tnum` (DESIGN.md, Typography; solution doc, Tabular-nums).
	 */
	const CATATAN_ANGKA = `${CATATAN} tabular-nums`;
	/**
	 * Modul: latar petak dengan garis rambut, dan hanya garis bawahnya yang
	 * digambar — dua modul bersebelahan berbagi satu garis, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const MODUL = 'border border-border border-b-0 bg-card';
	/**
	 * Jarak 8px sebelum modul Daftar bukan kelalaian: prototipe layar ini sendiri
	 * yang memisahkan kedua section dengan `margin-top: 8px`, jadi keduanya memang
	 * tidak bersebelahan — dan garis yang dihemat Shared-Hairline tetap dihemat di
	 * tempat yang berlaku, yaitu antara strip judul dan modul di bawahnya.
	 */
	const MODUL_DAFTAR = `${MODUL} mt-2`;
	/** Kepala modul: Wash Grey, ditutup garis tinta — satu-satunya penanda kepala. */
	const MODUL_KEPALA =
		'flex items-center justify-between gap-2 border-b border-foreground bg-muted px-2 py-1.5';
	const MODUL_JUDUL = 'text-[13px] font-bold tracking-[0.01em]';
	/** Satu blok isi di dalam modul, dan garis rambut bawahnya milik modul itu. */
	const MODUL_BARIS = 'border-b border-border px-2 py-1.5';
	const MODUL_CATATAN = `${MODUL_BARIS} text-xs`;
	/** Fields & Inputs: field 26px, label 12px/600 di atasnya dengan jarak 3px. */
	const FORMGRID = 'grid grid-cols-1 gap-x-3 gap-y-[10px] min-[901px]:grid-cols-3';
	const FIELD = 'flex flex-col gap-[3px]';
	const FIELD_LABEL = 'text-xs leading-[1.2] font-semibold';
	const FIELD_INPUT =
		'h-[26px] border-border bg-card px-1.5 py-0 text-[13px] shadow-none focus-visible:border-foreground aria-invalid:ring-0 md:text-[13px]';
	const FIELD_SELECT = cn(
		FIELD_INPUT,
		'w-full data-[size=default]:h-[26px] [&_svg:not([class*=size-])]:size-3'
	);
	/**
	 * Baris pesan — galat maupun hasil — di dalam modul ini, dan warnanya tinta,
	 * bukan merah utilitas. Merah hanya dua pekerjaan, tab dan harga (DESIGN.md,
	 * Secondary & Three-Percent Rule); yang memakai merah hanyalah garis dan
	 * outline field yang tidak valid, dan keputusan itu sudah berlaku seragam di
	 * layar terporting (#46).
	 */
	const PESAN = 'text-xs font-semibold';
	/**
	 * Satu baris pesan yang berdiri di luar modul, tepat di bawahnya: sisi dan
	 * bawahnya digambar, atasnya tidak — garis yang memisahkannya dari modul di
	 * atasnya sudah digambar modul itu, dan satu garis dipakai bersama, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const BARIS_GALAT = 'border border-t-0 border-border bg-card px-2 py-1.5';
	/** Baris aksi di bawah kisi field: `gap: 4px` dan `margin-top: 10px` prototipe. */
	const AKSI = 'mt-[10px] flex flex-wrap items-center gap-1';
	/**
	 * Commit Button: satu-satunya bidang bertinta penuh di formulir ini, dan saat
	 * mati ia kehilangan tintanya lalu jadi putih bergaris putus-putus — bukan
	 * pudar (DESIGN.md, Commit Button & State-Is-Not-Faded).
	 *
	 * DESIGN.md menggambarnya 40px selebar kolom keranjang; di sini ia berdiri di
	 * baris aksi di dalam modul, dan prototipe layar ini menulisnya `.btn--solid` —
	 * tinggi `.btn` dunia ini, 26px, tintanya tetap penuh. Yang diwarisi dari Commit
	 * Button adalah tintanya, bukan lebarnya, dan ukuran 26px itu kosakata yang sama
	 * dengan form Tambah Stok (#34).
	 */
	const COMMIT =
		'h-[26px] border-foreground bg-primary px-2 text-[13px] font-semibold text-primary-foreground hover:border-foreground hover:bg-primary hover:text-primary-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:border-dashed disabled:border-border disabled:bg-card disabled:text-foreground disabled:opacity-100';
	/**
	 * `.btn` 26px/13px, bukan `.tbl__acts .btn` 22px/12px: aksi di sini berdiri di
	 * dalam baris daftar, dan baris daftar memakai ukuran dunia ini — ukuran 22px
	 * hanya milik aksi di dalam sel tabel (prototipe, `.lirow__acts`).
	 *
	 * Dipakai dengan `variant="ghost"`: varian itu sudah membawa `hover:bg-muted`,
	 * persis latar hover dunia ini, dan tidak membawa apa pun yang harus dilawan
	 * (solution doc §3). Saat mati ia mempertahankan tintanya dan hanya kehilangan
	 * kursor — yang kehilangan tinta adalah tombol bertinta penuh.
	 */
	const AKSI_MODUL =
		'h-[26px] border-border bg-card px-2 text-[13px] font-semibold hover:border-foreground focus-visible:border-foreground disabled:pointer-events-auto disabled:cursor-not-allowed disabled:opacity-100';
	/**
	 * Daftar baris: satu garis rambut bersama antar baris, bukan dua
	 * (DESIGN.md, Shared-Hairline).
	 */
	const DAFTAR = 'divide-y divide-border border-b border-border';
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
	/**
	 * Quiet tag untuk keadaan Nonaktif: isian Wash Grey bergaris rambut, tinggi 15px
	 * (DESIGN.md, Tags), dan garis coretnya datang dari prototipe layar ini
	 * (`.tag--nonaktif { text-decoration: line-through }`). Prototipe tidak
	 * memudarkannya, dan opasitas memang dilarang sebagai penanda keadaan
	 * (DESIGN.md, State-Is-Not-Faded).
	 */
	const TAG_NONAKTIF =
		'inline-flex h-[15px] items-center border border-border bg-muted px-[5px] text-xs font-semibold line-through decoration-1 whitespace-nowrap';
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
					<Label for="pengguna-username" class={FIELD_LABEL}>Username</Label>
					<Input
						id="pengguna-username"
						name="username"
						autocomplete="off"
						class={FIELD_INPUT}
						bind:value={username}
						aria-invalid={fieldErrors.username ? true : undefined}
					/>
					{#if fieldErrors.username}
						<p class={PESAN}>{fieldErrors.username}</p>
					{/if}
				</div>

				<div class={FIELD}>
					<Label for="pengguna-password" class={FIELD_LABEL}>Password</Label>
					<Input
						id="pengguna-password"
						name="password"
						type="password"
						autocomplete="new-password"
						class={FIELD_INPUT}
						bind:value={password}
						aria-invalid={fieldErrors.password ? true : undefined}
					/>
					{#if fieldErrors.password}
						<p class={PESAN}>{fieldErrors.password}</p>
					{/if}
				</div>

				<div class={FIELD}>
					<Label for="pengguna-role" class={FIELD_LABEL}>Peran</Label>
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
				<Button class={COMMIT} type="submit" disabled={create.isPending}>
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
			<p class={MODUL_CATATAN}>Memuat daftar Pengguna…</p>
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
			<p class={MODUL_CATATAN}>Belum ada Pengguna lain. Tambahkan Kasir lewat formulir di atas.</p>
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
									<span class={TAG_NONAKTIF}>Nonaktif</span>
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
