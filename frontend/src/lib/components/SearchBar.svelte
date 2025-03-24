<script lang="ts">
	export let onSelect: (data: { name: string; rank: string }) => void = () => {};

	let searchTerm = '';
	let suggestions: any[] = [];
	let isLoading = false;

	let error: string | null = null;

	$: if (searchTerm.length >= 2) {
		fetchSuggestions(searchTerm);
	}

	async function fetchSuggestions(query: string) {
		isLoading = true;
		try {
			const res = await fetch(`http://localhost:8080/api/search?q=${encodeURIComponent(query)}`);
			if (res.ok) {
				suggestions = await res.json();
			} else {
				error = `API ERror: ${res.status}`;
			}
		} catch (err) {
			error = 'Search failed';
			console.error(err);
		}
		isLoading = false;
	}

	function selectSuggestion(suggestion: any) {
		searchTerm = `${suggestion.scientific_name} ${suggestion.authorship || ''} (${suggestion.rank})`;
		suggestions = [];

		onSelect({
			name: suggestion.scientific_name,
			rank: suggestion.rank
		});
	}
</script>

<style>
	.search-container {
		position: relative;
		width: 100%;
		max-width: 500px;
	}
	.suggestions {
		position: absolute;
		background: white;
		border: 5px solid #ccc;
		border-radius: 0.25rem;
		width: 100%;
		max-height: 200px;
		overflow-y: auto;
		z-index: 1000;
	}
	.suggestion {
		padding: 0.5rem;
		cursor: pointer;
	}
	.suggestions:hover {
		background: #f0f0f0;
	}
	.no-media {
		color:red;
	}
	i {
		font-style: italic;
	}
</style>

<div class="search-container">
	<input
		type="text"
		placeholder="Search by scientific name"
		bind:value={searchTerm}
		autocomplete="off"
	/>

	{#if isLoading}
		<p>Loading...</p>
	{/if}

	{#if suggestions.length > 0}
		<div class="suggestions">
			{#each suggestions as s}
				<div
					class="suggestion {s.has_media ? '' : 'no-media'}"
					on:click={() => selectSuggestion(s)}
				>
					{s.scientific_name}
					{#if s.authorship}
						<i> {s.authorship}</i>
					{/if}
					&nbsp;({s.rank})
				</div>
			{/each}
		</div>
	{/if}

	{#if error}
		<p style="color: red;">{error}</p>
	{/if}
</div>
