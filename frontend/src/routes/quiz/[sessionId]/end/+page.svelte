<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	let score = 0;
	let total = 0;
	let meta: any = null;

	onMount(() => {
		const $page = get(page);
		score = parseInt($page.url.searchParams.get('score') || '0');
		total = parseInt($page.url.searchParams.get('total') || '0');

		const metaRaw = $page.url.searchParams.get('meta');
		if (metaRaw) {
			meta = JSON.parse(decodeURIComponent(metaRaw));
		}
	});
</script>

<style>
	.end-screen {
		width: 100%;
		height: 100vh;
		background-image: url('/mm/bg.svg');
		background-size: cover;
		background-position: center;
		background-attachment: fixed;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		font-family: 'OldNewspaperTypes', serif;
		color: #ccc;
		gap: 2rem;
		text-align: center;
	}

	h1 {
		font-size: 4rem;
	}

    h3 {
        font-size: 2rem;
        font-weight: bold;
        color: #757575;
    }

	.quiz-summary {
		font-size: 1.75rem;
		background: rgba(0, 0, 0, 0.6);
		padding: 1.25rem 2rem;
		border-radius: 1rem;
		box-shadow: 0 0.5rem 1rem rgba(0,0,0,0.5);
	}

	.button {
		font-size: 2rem;
		cursor: pointer;
	    color: #595959;
        font-weight: bold;
        text-shadow:
            -0.4px -0.4px 0 #E8E8E8,
            0.4px -0.4px 0 #E8E8E8,
            -0.4px  0.4px 0 #E8E8E8,
            0.4px  0.4px 0 #E8E8E8;
	}

	.button:hover {
        color: #2D2929;
	}
</style>

<div class="end-screen">
	<h1>Quiz Complete</h1>
	<div class="quiz-summary">
		<h3>You scored: {score} / {total}</h3>
        <br />
		{#if meta}
			<p>Taxon: {meta.name}<br />Target Rank: {meta.targetRank}<br />Option Count: {meta.optionCount}</p>
		{/if}
	</div>

	<div class="button" on:click={() => goto('/quiz')}>Return to Quiz Setup</div>
</div>

