<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import Input from '$lib/components/design/Input.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import { exchangeAccessLink, getCurrentSession, logout, type CurrentSession } from '$lib/api/auth';

	let loading = true;
	let exchanging = false;
	let loggingOut = false;
	let pageError = '';
	let session: CurrentSession = { authenticated: false };
	let tokenInput = '';

	onMount(async () => {
		const urlToken = new URL(window.location.href).searchParams.get('token')?.trim() ?? '';
		if (urlToken) {
			tokenInput = urlToken;
			await doExchange(urlToken);
			return;
		}
		try {
			session = await getCurrentSession();
		} catch {
			session = { authenticated: false };
		} finally {
			loading = false;
		}
	});

	async function doExchange(token: string) {
		exchanging = true;
		pageError = '';
		try {
			session = await exchangeAccessLink(token);
			await goto('/tasks');
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not use that access link.';
			session = { authenticated: false };
		} finally {
			exchanging = false;
			loading = false;
		}
	}

	async function submitToken() {
		const token = tokenInput.trim();
		if (!token) return;
		await doExchange(token);
	}

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
		<h1 class="display">
			{session.authenticated ? 'You are signed in.' : 'Use your beta access link.'}
		</h1>

		{#if loading || exchanging}
			<InfoBox title={exchanging ? 'Checking your access link…' : 'Checking your session…'}>
				{exchanging
					? 'One moment while we exchange your access link.'
					: 'Looking up whether you are already signed in.'}
			</InfoBox>
		{:else if session.authenticated}
			<p class="lede">
				<strong>{session.user?.display_name}</strong> is signed in on this browser.
			</p>
			<InfoBox>
				This beta currently uses operator-issued access links rather than passwords.
			</InfoBox>
			<div class="actions">
				<Button variant="primary" on:click={() => goto('/tasks')}>Continue</Button>
				<Button variant="secondary" on:click={signOut} disabled={loggingOut}>
					{loggingOut ? 'Signing out…' : 'Sign out'}
				</Button>
			</div>
		{:else}
			<p class="lede">
				Open the one-time beta access link from the operator on this browser to sign in.
			</p>
			<InfoBox>
				If your last link expired or was already used, ask the operator for a new one.
			</InfoBox>
			<form class="token-form" on:submit|preventDefault={submitToken}>
				<Input
					id="access-token"
					label="Beta access token"
					placeholder="Paste your access link or token"
					required
					bind:value={tokenInput}
					error={pageError ? ' ' : ''}
				/>
				<Button variant="primary" type="submit" disabled={exchanging} fullWidth>
					{exchanging ? 'Signing in…' : 'Sign in'}
				</Button>
			</form>
		{/if}

		{#if pageError}
			<p class="error-banner" role="alert">{pageError}</p>
		{/if}
	</section>
</div>

<style>
	.page {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: var(--space-6) var(--space-5);
	}

	.card {
		width: 100%;
		max-width: var(--surface-max-width);
		padding: var(--space-8) var(--space-6);
		background: var(--paper);
		border: 1px solid var(--line);
		border-radius: var(--radius-3xl);
		box-shadow: var(--shadow-md);
	}

	.eyebrow {
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--primary-2);
		font-weight: 600;
		margin-bottom: var(--space-3);
	}

	.display {
		font-family: var(--font-display);
		font-size: 34px;
		line-height: 1.1;
		letter-spacing: -0.005em;
		color: var(--ink);
		font-weight: 400;
		margin-bottom: var(--space-4);
	}

	.lede {
		font-size: 15.5px;
		color: var(--ink-2);
		line-height: 1.6;
		margin-bottom: var(--space-4);
	}

	.token-form {
		display: grid;
		gap: var(--space-4);
		margin-top: var(--space-4);
	}

	.actions {
		display: flex;
		gap: var(--space-4);
		margin-top: var(--space-4);
		min-width: 0;
	}

	.error-banner {
		color: var(--rose);
		font-weight: 600;
		padding: var(--space-4);
		border-radius: var(--radius-lg);
		background: var(--rose-soft);
		margin-top: var(--space-4);
	}

	@media (max-width: 540px) {
		.card {
			padding: var(--space-6) var(--space-4);
			border-radius: var(--radius-2xl);
		}

		.actions {
			flex-direction: column;
		}
	}
</style>
