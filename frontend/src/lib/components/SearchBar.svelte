<script lang="ts">
	export let onSelect: (data: { name: string; rank: string }) => void = () => {};

	let searchTerm = '';
	let suggestions: any[] = [];
	let isLoading = false;

	let error = string | null = null;

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
			name = suggestion.scientific_name;
			rank = suggestion.rank;
		});
	}
</script>
