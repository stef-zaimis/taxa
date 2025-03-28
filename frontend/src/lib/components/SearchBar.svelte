<script lang="ts">
	export let onSelect: (data: { name: string; rank: string; authorship: string }) => void = () => {};

	let searchTerm = '';
	let suggestions: any[] = [];
	let isLoading = false;
	let error: string | null = null;
	let isFocused = false;

	let debounceTimeout: ReturnType<typeof setTimeout> | null = null;
	let blurTimeout: ReturnType<typeof setTimeout> | null = null;

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

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		searchTerm = target.value

		if (debounceTimeout) clearTimeout(debounceTimeout);

		if (searchTerm.length >= 2) {
			debounceTimeout = setTimeout(() => {
				fetchSuggestions(searchTerm);
			}, 1000);
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && searchTerm.length >= 2) {
			if (debounceTimeout) clearTimeout(debounceTimeout);
			fetchSuggestions(searchTerm);
		}
	}

	function handleFocus() {
		isFocused = true;
		if (blurTimeout) clearTimeout(blurTimeout);

		if (searchTerm.length >=2 && suggestions.length === 0) {
			fetchSuggestions(searchTerm);
		}
	}

	function handleBlur() {
		blurTimeout = setTimeout(() => {
			isFocused = false;
		}, 150);
	}

	function selectSuggestion(suggestion: any) {
		onSelect({
			name: suggestion.scientific_name,
			rank: suggestion.rank,
			authorship: suggestion.authorship || ''
		});
		isFocused = false;
	}
</script>

<style>
	.search-container {
		position: relative;
		width: 100%;
		max-width: 500px;
	}
	.suggestions {
		list-style: none;
		margin: 0;
		padding: 0;
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
		on:input={handleInput}
		on:keydown={handleKeydown}
		on:focus={handleFocus}
		on:blur={handleBlur}
		autocomplete="off"
	/>

	{#if isLoading}
		<p>Loading...</p>
	{/if}

	{#if isFocused && suggestions.length > 0}
		<ul class="suggestions">
			{#each suggestions as s}
				<li
					class="suggestion {s.has_media ? '' : 'no-media'}"
					role="option"
					tabindex="0"
					on:mousedown|preventDefault={() => selectSuggestion(s)}
				>
					{s.scientific_name}
					{#if s.authorship}
						<i> {s.authorship}</i>
					{/if}
					&nbsp;({s.rank})
				</li>
			{/each}
		</ul>
	{/if}

	{#if error}
		<p style="color: red;">{error}</p>
	{/if}
</div>
