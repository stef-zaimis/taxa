<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	let loading = false;
	let imageUrl = '';
	let options: any[] = [];
	let correctAnswer = '';
	let selectedAnswer: string | null = null;
	let resultText = '';
	let resultColor = '';


	let quizMeta: {
		rank: string;
		name: string;
		targetRank: string;
		optionCount: number;
	} | null = null;

	onMount(() => {
		const $pageData = get(page);
		const query = $pageData.url.searchParams.get('data');
		const sessionId = $pageData.params.sessionId;

		if (!query || !sessionId) return;

		const data = JSON.parse(decodeURIComponent(query));
		console.log('Received quiz data:', data);

		const metaRaw = sessionStorage.getItem(`quiz-meta-${sessionId}`);
		if (metaRaw) {
			quizMeta = JSON.parse(metaRaw);
		} else {
			console.error('No quizMeta found in sessionSTorage');
		}
	
		console.log('Parsed meta:', quizMeta);
		setQuestion(data);
	});

	function setQuestion(data: any) {
		imageUrl = data.imageUrl;
		options = data.options;
		correctAnswer = data.correctAnswer.scientificName;
		resultText = '';
		selectedAnswer = null;
	}

	async function fetchNextQuestion() {
		if (!quizMeta) {
			console.error("Missing quiz metadata");
			resultText = "Something went wrong.";
			resultColor = "Red";
			return;
		}

		loading = true;

		try {
			const { rank, name, targetRank, optionCount } = quizMeta;
			const url = `http://localhost:8080/api/quiz?rank=${rank}&name=${name}&targetRank=${targetRank}&optionCount=${optionCount}`;
			const res = await fetch(url);
			if (!res.ok) throw new Error("Failed to fetch next question");

			const data = await res.json();
			setQuestion(data);
		} catch (err) {
			console.error("Next question error:", err);
			resultText = "Something went wrong.";
			resultColor = "red";
		}
		loading = false;
	}

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
	<button on:click={() => handleClick(opt.scientificName)} disabled={selectedAnswer !== null}>
		{opt.scientificName}
	</button>
{/each}

{#if selectedAnswer}
	<p style="color: {resultColor}; font-weight: bold;">{resultText}</p>
	<button on:click={fetchNextQuestion} disabled={loading}>
		{loading ? 'Loading...' : 'Next Question'}
	</button>
{/if}

{#if loading}
	<p>Loading next question...</p>
{/if}

<br />
<a href="quiz"> Back to Selection</a>
