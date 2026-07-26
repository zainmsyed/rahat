<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { exchangeAccessLink, getCurrentSession, logout, type CurrentSession } from '$lib/api/auth';

	let loading = true;
	let exchanging = false;
	let loggingOut = false;
	let pageError = '';
	let session: CurrentSession = { authenticated: false };

	onMount(async () => {
		const token = new URL(window.location.href).searchParams.get('token')?.trim() ?? '';
		if (token) {
			exchanging = true;
			try {
				session = await exchangeAccessLink(token);
				await goto('/tasks');
				return;
			} catch (error) {
				pageError = error instanceof Error ? error.message : 'Could not use that access link.';
			} finally {
				exchanging = false;
			}
		}
		try {
			session = await getCurrentSession();
		} catch {
			session = { authenticated: false };
		} finally {
			loading = false;
		}
	});

	async function signOut() {
		pageError = '';
		loggingOut = true;
		try {
			await logout();
			session = { authenticated: false };
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not sign out.';
		} finally {
			loggingOut = false;
		}
	}
</script>

<svelte:head>
	<title>Rahat login</title>
</svelte:head>

<div class="page">
	<section class="card">
		<p class="eyebrow">Rahat beta</p>
		<h1>{session.authenticated ? 'You are signed in.' : 'Use your beta access link.'}</h1>
		{#if loading || exchanging}
			<p>{exchanging ? 'Checking your access link…' : 'Checking your session…'}</p>
		{:else if session.authenticated}
			<p><strong>{session.user?.display_name}</strong> is signed in on this browser.</p>
			<p>This beta currently uses operator-issued access links rather than passwords.</p>
			<div class="actions">
				<a href="/tasks">Continue</a>
				<button type="button" class="ghost" on:click={signOut} disabled={loggingOut}>
					{loggingOut ? 'Signing out…' : 'Sign out'}
				</button>
			</div>
		{:else}
			<p>Open the one-time beta access link from the operator on this browser to sign in.</p>
			<p>If your last link expired or was already used, ask the operator for a new one.</p>
		{/if}
		{#if pageError}
			<p class="error">{pageError}</p>
		{/if}
	</section>
</div>

<style>
	.page {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 2rem;
	}
	.card {
		max-width: 38rem;
		padding: 2rem;
		border-radius: 1.5rem;
		background: white;
		box-shadow: 0 18px 45px rgba(24, 34, 47, 0.08);
	}
	.eyebrow {
		margin: 0 0 0.75rem;
		font-size: 0.875rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #2a6df4;
	}
	.actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
		align-items: center;
	}
	.actions a,
	button {
		padding: 0.85rem 1.15rem;
		border-radius: 999px;
		text-decoration: none;
		font-weight: 600;
		border: none;
		background: #18222f;
		color: white;
		cursor: pointer;
	}
	button.ghost {
		background: white;
		color: #18222f;
		border: 1px solid #cbd5e1;
	}
	.error {
		margin-top: 1rem;
		color: #b42318;
	}
</style>
