<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	let imageUrl = '';
	let options: any[] = [];
	let correctAnswer = '';
	let selectedAnswer: string | null = null;
	let resultText = '';
	let resultColor = '';

	onMount(() => {
		const $pageData = get(page);
		const query = $page.url.searchParams.get('data');
		if (!query) return;

		const data = JSON.parse(decodeURIComponent(query));
		console.log('Received quiz data:', data);

		imageUrl = data.imageUrl;
		options = data.options;
		correctAnswer = data.correctAnswer.scientificName;
	});

	function handleClick(selected: string) {
		selectedAnswer = selected;
		if (selected === correctAnswer) {
			resultText = 'Correct!';
			resultColor = 'green';
		} else {
			resultText = `Incorrect, the correct answer was: ${correctAnswer}`;
			resultColor = 'red';
		}
	}
</script>

<style>
	.image {
		max-width: 100%;
		max-height: 600px;
		object-fit: contain;
		border-radius: 0.5rem;
		margin-bottom: 1rem;
		display: block;
		margin-left: auto;
		margin-right: auto;
	}
	button {
		display: block;
		margin: 0.5rem 0;
		padding: 0.5rem 1rem;
		font-size: 1.1rem;
	}
</style>

<h1>Quiz Question</h1>

{#if imageUrl}
	<img src={imageUrl} alt="Taxon image" class="image" />
{/if}

{#each options as opt}
	<button on:click={() => handleClick(opt.scientificName)}>
		{opt.scientificName}
	</button>
{/each}

{#if selectedAnswer}
	<p style="color: {resultColor}; font-weight: bold;">{resultText}</p>
{/if}
