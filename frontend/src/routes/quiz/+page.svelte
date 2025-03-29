<script lang="ts">
	import { goto } from '$app/navigation';
	import SearchBar from '$lib/components/SearchBar.svelte';

	let rank = '';
	let authorship = '';
	let name = '';
	let targetRank = '';
	let optionCount = 4;

	let result: any = null;
	let error: string | null = null;

	function generateSessionId() {
		return crypto.randomUUID();
	}

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
			const data = await res.json();
			const encoded = encodeURIComponent(JSON.stringify(data));
			console.log(JSON.stringify(data));

			const sessionId = generateSessionId();
			sessionStorage.setItem(`quiz-${sessionId}`, JSON.stringify(data));
			goto(`/quiz/${sessionId}?data=${encoded}`);
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
	<SearchBar mode="taxon" onSelect={({ name: selectedName, rank: selectedRank, authorship: selectedAuthorship }) => {
		name = selectedName;
		rank = selectedRank;
		authorship = selectedAuthorship;
	}} placeholder="Search for taxon (e.g. Animalia)" />
	
	<input placeholder="Scientific Name (e.g. Animalia)" bind:value={name} />
	<input placeholder="Authorship (e.g. Linnaeus, 1758)" bind:value={authorship} />
	<SearchBar mode="rank" onSelect={({ name: selectedRank }) => { rank = selectedRank; }} placeholder="Search for rank (e.g. Kingdom)" />
	<SearchBar mode="rank" onSelect={({ name: selectedTargetRank}) => { targetRank = selectedTargetRank; }} placeholder="Search for target rank (e.g. Order)" />
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
