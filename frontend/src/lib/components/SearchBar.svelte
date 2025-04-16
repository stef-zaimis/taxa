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
		console.log('VITE_API_URL:', import.meta.env.VITE_API_URL);
		try {
			const baseUrl = import.meta.env.VITE_API_URL;
			const endpoint = mode === 'rank' ? '/search/ranks' : '/search/taxa';
			const res = await fetch(`${baseUrl}${endpoint}?q=${encodeURIComponent(query)}`);
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
		searchTerm = mode === 'taxon' ? `${suggestion.scientific_name}${suggestion.authorship ? ` ${suggestion.authorship}` : ''}` : suggestion;
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
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		width: 100%;
	}

	.searchbar-input {
		width: 100%;
		height: 100%;
		background: transparent;
		border: none;
		outline: none;
		box-shadow: none;
		color: black;
		font-family: 'OldNewspaperTypes', serif;
		font-size: 1.5rem;
		z-index: 1;
		text-align: center;
	}

	.searchbar-input::placeholder {
		text-align: center;
		font-size: 1.5rem;
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
		top: 100%;
		left: 0;
		width: 100%;
		max-height: 200px;
		font-family: 'OldNewspaperTypes', serif !important;
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
	.loading-message {
		font-family: 'OldNewspaperTypes', serif !important;
		position: absolute;
		top: 60%;
		left: 0;
		width: 100%;
		margin-top: 0.25rem;
		text-align: center;
		color: black;
		z-index: 999;
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
		class="searchbar-input"
	/>

	{#if isLoading}
		<p class="loading-message">Loading...</p>
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
		<p class="loading-message" style="color: red;">{error}</p>
	{/if}
</div>
