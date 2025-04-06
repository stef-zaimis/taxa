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
	let hintActive = false;


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
	:root {
		font-size: 16px; /* 1rem = 16px (default), easily scalable */
	}

	.quiz-container {
		width: 100vw;
		min-height: 100vh;
		background: url('/quiz/bg.webp');
		background-size: cover;
		background-position: center;
		position: relative;
		background-attachment: fixed;
		background-repeat: no-repeat;
		background-color: black;
		overflow: hidden;
		padding: 1rem;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
	}

	.hud-placeholder {
		position: absolute;
		top: 1rem;
		right: 1rem;
		width: auto;
		height: auto;
		z-index: 10;
		height: 6rem; /* ~96px */
		flex-shrink: 0;
	}

	.main-content {
		flex: 1;
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		gap: 2.5rem; 
		position: relative;
		box-sizing: border-box;
		height: 100vh;
	}

	.content-core {
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		gap: 2.5rem;
		margin-right: 25rem;
	}

	.image-frame {
		position: relative;
		width: min(90vw, 55rem); /* Cap at 480px */
		aspect-ratio: 1 / 1;
		flex-shrink: 0;
	}

	.quiz-image {
		position: absolute;
		top: 5%;
		left: 5%;
		width: 90%;
		height: 90%;
		object-fit: contain;
		z-index: 1;
	}

	.hint-icon {
		position: absolute;
		top: 0rem;
		right: 0rem;
		width: 11rem;
		height: 11rem;
		transform: translate(20%, -20%);
		cursor: pointer;
		z-index: 3;
	}

	.frame-overlay {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		z-index: 2;
		pointer-events: none;
	}

	.options-container {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		flex-shrink: 1;
		flex-grow: 1;
		min-width: 16rem;
		max-width: 28rem;
	}

	.option-row {
		display: flex;
		align-items: center;
		gap: 0.75rem; /* 12px */
	}

	.die-icon {
		width: 5rem;
		height: 5rem;
	}

	.option-panel {
		position: relative;
		width: 25rem;
		height: 6rem;
		cursor: pointer;
	}

	.option-text {
		position: absolute;
		top: 48%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 2;
		color: black;
		font-size: 1.5rem;
		font-weight: bold;
		text-align: center;
		width: 90%;
		pointer-events: none;
		line-height: 1.2;
	}

	.panel-bg {
		width: 100%;
		height: 100%;
		object-fit: contain;
		object-position: center;
		z-index: 1;
		position: absolute;
		top: 0;
		left: 0;
	}

	.navigation-buttons {
		position: absolute;
		right: 2rem;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		gap: 1.5rem;
	}

	.nav-button {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		width: 8rem;
		height: 8rem;
	}

	.nav-button:disabled img {
		opacity: 0.4;
		pointer-events: none;
	}

	.result-text {
		position: absolute;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%);
		background-color: white;
		padding: 0.75rem 1.5rem;
		border-radius: 0.75rem;
		font-size: 1.2rem;
		font-weight: bold;
		box-shadow: 0 0.25rem 0.375rem rgba(0, 0, 0, 0.2);
		white-space: nowrap;
	}

	@media (max-width: 768px) {
		.main-content {
			flex-direction: column;
			align-items: center;
		}

		.navigation-buttons {
			position: static;
			flex-direction: row;
			justify-content: center;
			margin-top: 2rem;
			transform: none;
		}
	}
</style>

<div class="quiz-container">
	<div class="hud-placeholder"></div>

	<div class="main-content">
		<div class="content-core">
			<div class="image-frame">
				{#if imageUrl}
					<img src={imageUrl} alt="Taxon image" class="quiz-image" />
				{/if}
				<img class="hint-icon" src={hintActive ? '/quiz/lightbulb_on.webp' : '/quiz/lightbulb_off.webp'} alt="Hint" on:click={() => hintActive = !hintActive} />
				
				<img class="frame-overlay" src="/quiz/frame.webp" alt="Frame" />
			</div>
			
			<div class="options-container">
				{#each options.slice(0,6) as opt, i}
					<div class="option-row">
						<img src={`/quiz/dice/die_${i+1}.webp`} class="die-icon" />
						<div class="option-panel" on:click={() => handleClick(opt.scientificName)} class:selected={selectedAnswer === opt.scientificName}>
							<span class="option-text">{opt.scientificName}</span>
							<img class="panel-bg" src="/quiz/option_panel.webp" alt="Option panel">
						</div>
					</div>
				{/each}

				{#if options.length > 6}
					{#each options.slice(6) as opt}
						<div class="option-row">
							<div class="option-panel" on:click={() => handleClick(opt.scientificName)} class:selected={selectedAnswer === opt.scientificName}>
								<span class="option-text">{opt.scientificName}</span>
								<img class="panel-bg" src="/quiz/option_panel.webp" alt="Option panel">
							</div>
						</div>
					{/each}
				{/if}
			</div>
		</div>

		<div class="navigation-buttons">
			<button class="nav-button back" disabled>
				<img src="/quiz/left_arrow.webp" />
			</button>
			<button class="nav-button forward" on:click={fetchNextQuestion} disabled={loading || !selectedAnswer}>
				<img src="/quiz/right_arrow.webp" />
			</button>
		</div>
	</div>

		{#if selectedAnswer}
			<p class="result-text" style="color: {resultColor};">{resultText}</p>
	{/if}
</div>
