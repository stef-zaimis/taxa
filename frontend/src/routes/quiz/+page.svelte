<script lang="ts">
	import { goto } from '$app/navigation';
	import SearchBar from '$lib/components/SearchBar.svelte';

	let rank = '';
	let authorship = '';
	let name = '';
	let targetRank = '';
	let optionCount = null;
	let questionCount = null;
	
	let loading = false;

	let result: any = null;
	let error: string | null = null;

	function generateSessionId() {
		return crypto.randomUUID();
	}

	async function fetchQuiz() {
		error = null;
		result = null;
		loading = true;

		if (optionCount === null) {
			optionCount = 0;
		}

		const params = new URLSearchParams({
			rank,
			name,
			targetRank,
			optionCount: optionCount.toString()
		});

		try {
			const res = await fetch(`${import.meta.env.VITE_API_URL}/quiz?${params.toString()}`);
			if (!res.ok) {
				throw new Error(`API error: ${res.status}`);
			}
			const data = await res.json();

			const sessionId = generateSessionId();

			const quizMeta = { ...Object.fromEntries(params), questionCount: questionCount?.toString() || null };

			console.log('Saving quiz-meta:', quizMeta);
			sessionStorage.setItem(`quiz-meta-${sessionId}`, JSON.stringify(quizMeta));

			const dataWithMeta = {
				...data,
				rank,
				name,
				targetRank,
				optionCount
			};

			const encoded = encodeURIComponent(JSON.stringify(dataWithMeta));
			sessionStorage.setItem(`quiz-${sessionId}`, JSON.stringify(dataWithMeta));
			goto(`/quiz/${sessionId}?data=${encoded}`);
		} catch (err: any) {
			error = err.message || 'Something went wrong';
			console.error('Fetch error:', err);
		} finally {
			loading = false;
		}
	}
</script>

<style>
	@font-face {
		font-family: 'OldNewspaperTypes';
		src: url('/fonts/OldNewspaperTypes.ttf') format('truetype');
		font-weight: normal;
		font-style: normal;
	}

	h1, body {
		color: #ccc;
		font-family: 'OldNewspaperTypes', serif;
	}

	h1 {
		font-size: 5rem;
	}

	input,
	button,
	::placeholder,
	.panel-content input,
	.search-container input,
	.suggestion {
		font-family: 'OldNewspaperTypes', serif !important;
	}

	input:focus,
	.search-container input:focus {
		outline: none;
		box-shadow: none;
		border: none;
	}

	.all-content {
		width: 100%;
		height: 100vh;
		overflow: hidden;
		background-image: url('/mm/bg.svg');
		background-size: cover;
		background-position: center;
		background-attachment: fixed;
		position: relative;
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 7rem;
	}

	.top-popup {
		position: absolute;
		top: 2rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 999;
		padding: 1.25rem 2.5rem;
		border-radius: 1rem;
		font-size: 2.25rem;
		font-weight: bold;
		font-family: 'OldNewspaperTypes', serif;
		background-color: rgba(255, 255, 255, 0.95);
		box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.25);
		white-space: nowrap;
		text-align: center;
		pointer-events: none;
		animation: fadeInTop 0.2s ease-out;
	}

	.loading-popup {
		color: black;
	}

	.error-popup {
		color: red;
		border: 2px solid red;
	}

	@keyframes fadeInTop {
		from {
			opacity: 0;
			transform: translateX(-50%) translateY(-1rem);
		}
		to {
			opacity: 1;
			transform: translateX(-50%) translateY(0);
		}
	}

	.inputs {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 3rem;
	}

	.selection-wrapper {
		position: relative;
		width: 100%;
		max-width: 1000px;
		marg: 2rem auto;
		aspect-ratio: 1404 / 269;
	}

	.panel-bg {
		width: 100%;
		height: auto;
		display: block;
		border-radius: 12px;
	}

	.panel-content {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		display: flex;
		justify-content: space-evenly;
		height: 100%;
		width: 100%;
		align-items: center;
		padding: 0.5rem 1rem;
		box-sizing: border-box;
	}

	.panel-content input,
	.panel-content :global(.search-container input) {
		width: 90%;
		height: 60px;
		background-size: cover;
		background-position: center;
		color: black;
		font-size: 1.5rem;
		padding: 0.75rem;
		border-radius: 8px;
		border: none;
	}

	.input-panel {
		position: relative;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.input-panel img {
		position: relative;
		height: 100%;
		width: auto;
		z-index: 0;
		object-fit: contain;
		display: block;
		border-radius: 8px;
	}

	.input-panel input {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		background: transparent;
		border: none;
		padding: 0.5rem 1rem;
		color: black;
		font-family: 'OldNewspaperTypes', serif;
		font-size: 1rem;
		z-index: 1;
	}

	.input-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1;
	}

	.input-wrapper {
		width: 90%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.input-wrapper input {
		width: 100%;
		height: 100%;
		background: transparent;
		border: none;
		outline: none;
		color: black;
		font-family: 'OldNewspaperTypes', serif;
		font-size: 1.5rem;
		text-align: center;
	}

	button {
		margin: 0.5rem;
		padding: 0.5rem;
		font-size: 1.5rem;
		color: black;
		border: 1px solid #ccc;
		cursor: pointer;
		background-color: #eee;
		border-radius: 8px;
	}

	button:hover {
		background-color: #bbb;
	}
</style>

<div class="all-content">
	{#if loading}
		<div class="top-popup loading-popup">
			Loading...
		</div>
	{:else if error}
		<div class="top-popup error-popup">
			Error loading question
		</div>
	{/if}

	<h1>Create Your Quiz</h1>
	<div class="inputs">
		<div class="selection-wrapper">
			<img src="/selection/wooden_board.png" alt="Wooden Panel" class="panel-bg" />

			<div class="panel-content">
				<div class="input-panel"> 
					<img src="/selection/taxon_panel.png" alt="Taxon Input" />

					<div class="input-overlay">
						<SearchBar mode="taxon" onSelect={({ name: selectedName, rank: selectedRank, authorship: selectedAuthorship }) => {
							name = selectedName;
							rank = selectedRank;
							authorship = selectedAuthorship;
						}} placeholder="Taxon (e.g. Animalia)" />
					</div>
				</div>
				
				<div class="input-panel"> 
					<img src="/selection/taxonomic_level_panel.png" alt="Taxon Input" />

					<div class="input-overlay">
						<SearchBar mode="rank" onSelect={({ name: selectedTargetRank}) => { targetRank = selectedTargetRank; }} placeholder="Target Rank (e.g. Order)" />
					</div>
				</div>
				
				<div class="input-panel">
					<img src="/selection/option_count_panel.png" alt="Taxon Input" />

					<div class="input-overlay">
						<div class="input-wrapper">
							<input type="number" min="2" max="20" placeholder="Options" bind:value={optionCount} />
						</div>
					</div>
				</div>
				<div class="input-panel">
					<img src="/selection/option_count_panel.png" alt="Taxon Input" />
					<div class="input-overlay">
						<div class="input-wrapper">
							<input type="number" min="1" max="50" placeholder="Questions" bind:value={questionCount} />
						</div>
					</div>
				</div>
			</div>
		</div>
		<button on:click={fetchQuiz}>Get Quiz Question</button>
	</div>
</div>

{#if error}
	<p style="color: red;">Error: {error}</p>
{:else if result}
	<h2>Result:</h2>
	<pre>{JSON.stringify(result, null, 2)}</pre>
{/if}
