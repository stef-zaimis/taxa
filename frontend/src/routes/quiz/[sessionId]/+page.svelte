<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import { goto } from '$app/navigation';

	let loading = false;
	let imageUrl = '';
	let options: any[] = [];
	let correctAnswer = '';
	let selectedAnswer: string | null = null;
	let resultText = '';
	let resultColor = '';
	let hintActive = false;

	let imageClass = '';

	let locked = false
	let score = 0;
	let totalQuestions = 0;

	let quizMeta: {
		rank: string;
		name: string;
		targetRank: string;
		optionCount: number;
		questionCount?: string | null;
	} | null = null;

	let questionCount = null;

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
			questionCount = quizMeta?.questionCount ? parseInt(quizMeta.questionCount) : null;
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
		locked = false;
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
			const baseUrl = import.meta.env.VITE_API_URL;
			const url = `${baseUrl}/quiz?rank=${encodeURIComponent(rank)}&name=${encodeURIComponent(name)}&targetRank=${encodeURIComponent(targetRank)}&optionCount=${optionCount}`;
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
		if (locked || selectedAnswer) return;
	
		locked = true;
		selectedAnswer = selected;
		totalQuestions += 1;
	
		if (selected === correctAnswer) {
			score += 1;
			resultText = 'Correct!';
			resultColor = 'green';
		} else {
			resultText = `Incorrect, the correct answer was: ${correctAnswer}`;
			resultColor = 'red';
		}
	}

	function getOptionClass(optName: string) {
		if (!selectedAnswer) return '';

		const isCorrect = optName === correctAnswer;
		const isSelected = optName === selectedAnswer;

		return [
			isCorrect ? 'correct' : '',
			!isCorrect && isSelected ? 'incorrect' : '',
			isSelected ? 'selected' : ''
		].join(' ');
	}

	function handleImageLoad(event: Event) {
		const img = event.target as HTMLImageElement;
		if (!img) return;

		const { naturalWidth, naturalHeight } = img;

		if (naturalWidth === naturalHeight) {
			imageClass = '';
		} else if (naturalWidth > naturalHeight) {
			imageClass = 'border-top-bottom';
		} else {
			imageClass = 'border-left-right';
		}
	}
</script>

<style>
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
		overflow: auto;
		padding: 1%;
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
		height: auto;
		flex-wrap: wrap;
	}

	.top-popup {
		position: absolute;
		top: 1.5rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 999;
		padding: 1rem 2rem;
		border-radius: 0.75rem;
		font-size: 2rem;
		font-weight: bold;
		background-color: rgba(255, 255, 255, 0.9);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
		white-space: nowrap;
		text-align: center;
		pointer-events: none;
		animation: fadeIn 0.2s ease-out;
	}

	.loading-popup {
		color: black;
	}

	.error-popup {
		color: red;
		background-color: rgba(255, 255, 255, 0.95);
		border: 2px solid red;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateX(-50%) translateY(-1rem);
		}
		to {
			opacity: 1;
			transform: translateX(-50%) translateY(0);
		}
	}

	.content-core {
		display: flex;
		flex-direction: row;
		justify-content: center;
		align-items: center;
		gap: 2.5rem;
		flex-wrap: wrap;
	}

	.image-frame {
		position: relative;
		width: clamp(20rem, 48vw, 58rem);
		aspect-ratio: 1 / 1;
		flex-shrink: 1;
		flex-grow: 0;
	}

	.frame-bg {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		z-index: 0;
	}

	.frame-blur {
		position: absolute;
		top: 5%;
		left: 5%;
		width: 90%;
		height: 90%;
		z-index: 0;
		backdrop-filter: blur(5px);
		-webkit-backdrop-filter: blur(5px);
		border-radius: 0.5rem;
		background-color: black;
		pointer-events: none;
	}

	.quiz-image {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
		z-index: 1;
	}

	.image-wrapper {
		position: absolute;
		top: 5%;
		left: 5%;
		width: 90%;
		height: 90%;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: border 0.2s ease-in-out;
	}

	.quiz-image.border-top-bottom {
		border-top: 4px solid rgba(0, 0, 0, 1);
		border-bottom: 4px solid rgba(0, 0, 0, 1);
		border-left: none;
		border-right: none;
	}
	
	.quiz-image.border-left-right {
		border-left: 4px solid rgba(0, 0, 0, 1);
		border-right: 4px solid rgba(0, 0, 0, 1);
		border-top: none;
		border-bottom: none;
	}
	.hint-icon {
		position: absolute;
		top: 0;
		right: 0;
		width: 22%;
		aspect-ratio: 1;
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
		align-items: center;
		flex-shrink: 1;
		flex-grow: 1;
		min-width: 16rem;
		max-width: 28rem;
		box-sizing: border-box;
	}

	.option-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.die-icon {
		width: clamp(3rem, 6vw, 5rem);
		height: clamp(3rem, 6vw, 5rem);
	}

	.option-panel {
		position: relative;
		width: clamp(14rem, 35vw, 25rem);
		height: clamp(3.5rem, 8vw, 6rem);
		cursor: pointer;
	}

	.option-text {
		position: absolute;
		top: 48%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 2;
		color: black;
		font-size: clamp(1rem, 3vw, 1.5rem);
		font-weight: normal;
		text-align: center;
		width: 90%;
		pointer-events: none;
		line-height: 1.2;
		transition: color 0.3 ease, font-weight 0.2 ease;
	}

	.option-text.correct {
		color: green;
		font-weight: bold;
	}

	.option-text.incorrect {
		color: red;
	}

	.option-text.selected {
		font-weight: bold;
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
		right: 0%;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	@media (max-width: 1280px) {
		.navigation-buttons {
			position: static;
			flex-direction: row;
			justify-content: center;
			transform: none;
			margin-top: 2rem;
		}

		.main-content {
			flex-direction: column;
			align-items: center;
		}
	}

	.nav-button {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		width: clamp(4rem, 6vw, 8rem);
		height: clamp(4rem, 6vw, 8rem);
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
		z-index: 20;
		padding: 0.75rem 1.5rem;
		border-radius: 0.75rem;
		font-size: 1.2rem;
		font-weight: bold;
		box-shadow: 0 0.25rem 0.375rem rgba(0, 0, 0, 0.2);
		white-space: nowrap;
	}

</style>

<div class="quiz-container">
	{#if loading}
		<div class="top-popup loading-popup">
			Loading...
		</div>
	{:else if resultText === "Something went wrong."}
		<div class="top-popup error-popup">
			Error loading question
		</div>
	{/if}

	<div class="hud-placeholder">
		<div style="font-size: 2rem; font-weight: bold; color: white;">
			Score: {score} / {totalQuestions}
		</div>
	</div>

	<button class="return-button" on:click={() => goto('/quiz')}>Return to the selection screen</button>
	
	<div class="main-content">
		<div class="content-core">
			<div class="image-frame">
				<img src='/quiz/frame-fill-fabric-super-dark.png' alt="Frame Background" class="frame-bg"/>
	
				{#if imageUrl}
					<div class="image-wrapper">
						<img src={imageUrl} alt="Taxon image" class={`quiz-image ${imageClass}`} on:load={handleImageLoad} />
					</div>
				{/if}
				<img class="hint-icon" src={hintActive ? '/quiz/lightbulb_on.webp' : '/quiz/lightbulb_off.webp'} alt="Hint" on:click={() => hintActive = !hintActive} />
				
				<img class="frame-overlay" src="/quiz/frame.webp" alt="Frame" />
			</div>
			
			<div class="options-container">
				{#each options.slice(0,6) as opt, i}
					<div class="option-row"> <img src={`/quiz/dice/die_${i+1}.webp`} class="die-icon" />
						<div class="option-panel" on:click={() => handleClick(opt.scientificName)} class:selected={selectedAnswer === opt.scientificName}>
							<span class={"option-text " + getOptionClass(opt.scientificName)}>{opt.scientificName}</span>
							<img class="panel-bg" src="/quiz/option_panel.webp" alt="Option panel">
						</div>
					</div>
				{/each}

				{#if options.length > 6}
					{#each options.slice(6) as opt}
						<div class="option-row">
							<div class="option-panel" on:click={() => handleClick(opt.scientificName)} class:selected={selectedAnswer === opt.scientificName}>
								<span class={"option-text " + getOptionClass(opt.scientificName)}>{opt.scientificName}</span>
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
			<button class="nav-button forward" on:click={fetchNextQuestion} disabled={loading || !selectedAnswer || (questionCount !== null && totalQuestions >= questionLimit)}>
				<img src="/quiz/right_arrow.webp" />
			</button>
		</div>
	</div>

		{#if selectedAnswer}
			<p class="result-text" style="color: {resultColor};">{resultText}</p>
	{/if}
</div>
