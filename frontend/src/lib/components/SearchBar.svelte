<script lang="ts">
	export let onSelect: (data: { name: string; rank: string; authorship: string }) => void = () => {};
	export let mode: 'taxon' | 'rank' = 'taxon';
	export let placeholder: string = 'Search...';

	let searchTerm = '';
	let suggestions: any[] = [];
	let isLoading = false;
	let error: string | null = null;
	let isFocused = false;
	let minLength = mode == 'rank' ? 1 : 2;

	let debounceTimeout: ReturnType<typeof setTimeout> | null = null;
	let blurTimeout: ReturnType<typeof setTimeout> | null = null;

	async function fetchSuggestions(query: string) {
		isLoading = true;
		try {
			const endpoint = mode === 'rank' ? '/api/search/ranks' : '/api/search/taxa';
			const res = await fetch(`http://localhost:8080${endpoint}?q=${encodeURIComponent(query)}`);
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
		suggestions = [];
		const target = event.target as HTMLInputElement;
		searchTerm = target.value

		if (debounceTimeout) clearTimeout(debounceTimeout);

		if (searchTerm.length >= minLength) {
			debounceTimeout = setTimeout(() => {
				fetchSuggestions(searchTerm);
			}, 800);
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && searchTerm.length >= minLength) {
			if (debounceTimeout) clearTimeout(debounceTimeout);
			fetchSuggestions(searchTerm);
		}
	}

	function handleFocus() {
		isFocused = true;
		if (blurTimeout) clearTimeout(blurTimeout);

		if (searchTerm.length >=minLength && suggestions.length === 0) {
			fetchSuggestions(searchTerm);
		}
	}

	function handleBlur() {
		blurTimeout = setTimeout(() => {
			isFocused = false;
		}, 150);
	}

	function selectSuggestion(suggestion: any) {
		onSelect(mode === 'taxon' ? {
			name: suggestion.scientific_name,
			rank: suggestion.rank,
			authorship: suggestion.authorship || ''
		} : {
			name: suggestion,
			rank: suggestion,
			authorship: ''
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

	.search-container input {
		width: 100%;
		padding: 0.5rem;
		font-size: 1rem;
		color: black;
		background-color: white;
		border: 1px solid #ccc;
		border-radiu: 4px;
	}

	.suggestions {
		list-style: none;
		margin: 0;
		padding: 0;
		position: absolute;
		background: white;
		border: 1px solid #ccc;
		color: black;
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
	.suggestion:hover {
		background-color: #eee;
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
		placeholder={placeholder}
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
					class="suggestion {mode === 'taxon' && !s.has_media ? 'no-media' : ''}"
					role="option"
					tabindex="0"
					on:mousedown|preventDefault={() => selectSuggestion(s)}
				>
					{#if mode === 'rank'} 
						{s}
					{:else}
						<span>
							{s.scientific_name}
							{#if s.authorship}
								<i> {s.authorship}</i>
							{/if}
							&nbsp;({s.rank})
						</span>
					{/if}	
				</li>
			{/each}
		</ul>
	{/if}

	{#if error}
		<p style="color: red;">{error}</p>
	{/if}
</div>
