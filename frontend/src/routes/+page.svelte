<script lang="ts">
	import SearchBar from '$lib/components/SearchBar.svelte';

	let rank = '';
	let authorship = '';
	let name = '';
	let targetRank = '';
	let optionCount = 4;

	let result: any = null;
	let error: string | null = null;

	async function fetchQuiz() {
		error = null;
		result = null;

		const params = new URLSearchParams({
			rank,
			name,
			targetRank,
			optionCount: optionCount.toString()
		});

		try {
			const res = await fetch(`http://localhost:8080/api/quiz?${params.toString()}`);
			if (!res.ok) {
				throw new Error(`API error: ${res.status}`);
			}
			result = await res.json();
			console.log('Quiz response:', result);
		} catch (err: any) {
			error = err.message || 'Something went wrong';
			console.error('Fetch error:', err);
		}
	}
</script>

<style>
	input, button {
		margin: 0.5rem;
		padding: 0.5rem;
		font-size: 1rem;
	}
</style>

<h1>Taxa Quiz Generator</h1>

<div>
	<SearchBar onSelect={({ name: selectedName, rank: selectedRank, authorship: selectedAuthorship }) => {
		name = selectedName;
		rank = selectedRank;
		authorship = selectedAuthorship;
	}} />
	
	<input placeholder="Scientific Name (e.g. Animalia)" bind:value={name} />
	<input placeholder="Authorship (e.g. Linnaeus, 1758)" bind:value={authorship} />
	<input placeholder="Rank (e.g. Kingdom)" bind:value={rank} />
	<input placeholder="Target Rank (e.g. Order)" bind:value={targetRank} />
	<input type="number" min="2" max="20" placeholder="Option Count" bind:value={optionCount} />
	<br />
	<button on:click={fetchQuiz}>Get Quiz Question</button>
</div>

{#if error}
	<p style="color: red;">Error: {error}</p>
{:else if result}
	<h2>Result:</h2>
	<pre>{JSON.stringify(result, null, 2)}</pre>
{/if}
