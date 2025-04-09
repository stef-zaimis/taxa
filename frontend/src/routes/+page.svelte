<script lang="ts">
	import { goto } from '$app/navigation';

	function startQuiz() {
		goto('/quiz');
	}

	function goToOptions() {
		// TODO
	}

	function goToSettings() {
		// TODO
	}

	function exitGame() {
		// TODO
	}

	// Use the previous working logic:
	function handleMouseDown(event: MouseEvent) {
		const img = (event.currentTarget as HTMLElement).querySelector('.menu-img');
		if (img) {
			// Remove any existing bounce class so that :active works as expected
			img.classList.remove('bounce');
		}
	}

	function handleMouseUp(event: MouseEvent) {
		const img = (event.currentTarget as HTMLElement).querySelector('.menu-img');
		if (img) {
			// Trigger bounce animation after mouseup
			img.classList.add('bounce');
			setTimeout(() => img.classList.remove('bounce'), 250);
		}
	}
</script>

<div class="main-menu-wrapper">
	<div class="menu-wrapper-inner">
		<div class="menu-content">
			<img src="/mm/title.svg" alt="TAXA Title" class="title-img" draggable="false" />

			<button
				on:mousedown={handleMouseDown}
				on:mouseup={handleMouseUp}
				on:click={startQuiz}
				type="button"
				class="menu-button menu-button--start"
			>
				<img src="/mm/start_button.svg" alt="Start Button" class="menu-img" draggable="false" />
			</button>

			<div class="button-group">
				<button
					on:mousedown={handleMouseDown}
					on:mouseup={handleMouseUp}
					on:click={goToOptions}
					type="button"
					class="menu-button"
				>
					<img src="/mm/options_button.svg" alt="Options" class="menu-img" draggable="false" />
				</button>

				<button
					on:mousedown={handleMouseDown}
					on:mouseup={handleMouseUp}
					on:click={goToSettings}
					type="button"
					class="menu-button"
				>
					<img src="/mm/settings_button.svg" alt="Settings" class="menu-img" draggable="false" />
				</button>

				<button
					on:mousedown={handleMouseDown}
					on:mouseup={handleMouseUp}
					on:click={exitGame}
					type="button"
					class="menu-button"
				>
					<img src="/mm/exit_game_button.svg" alt="Exit" class="menu-img" draggable="false" />
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	.main-menu-wrapper {
		width: 100vw;
		min-height: 100vh;
		overflow: hidden;
		background-image: url('/mm/bg.svg');
		background-size: cover;
		background-position: center;
		background-attachment: fixed;
		position: relative;
		background-color: black;
	}

	.menu-wrapper-inner {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		min-height: 100vh;
		padding: 2rem 1rem;
		box-sizing: border-box;
	}

	.menu-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: clamp(1vh, 2.5vh, 3rem);
		width: 100%;
		max-width: 100%;
	}

	.title-img {
		width: 95vw;
		max-width: 77rem;
		max-height: 60vh;
		height: auto;
		margin-top: clamp(1vh, 4vh, 3rem);
		display: block;
		filter: drop-shadow(0 0 1.25rem rgba(255, 255, 255, 0.25));
	}

	.button-group {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 1.5rem;
		padding-top: clamp(1vh, 5vh, 4rem);
	}

	@media (min-width: 768px) {
		.button-group {
			gap: 3rem;
		}
	}

	/* -------- BUTTON STRUCTURE ---------------*/
	.menu-button {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		display: inline-block;
	}

	.menu-img {
		height: clamp(3.2vh, 3.6rem, 6vh);
		width: auto;
		max-width: 80vw;
		display: block;
		object-fit: contain;
		transition: transform 0.25s ease;
		will-change: transform;
	}

	/* -------- HOVER BEHAVIOR ---------------*/
	/* Default buttons shrink on hover */
	.menu-button:hover .menu-img {
		transform: scale(0.9);
	}

	/* Start button grows on hover */
	.menu-button--start:hover .menu-img {
		transform: scale(1.15);
	}

	/* -------- ACTIVE / CLICKED STATE ---------------*/
	/* When the button is pressed, we override the hover transform */
	.menu-button:active .menu-img {
		transform: scale(1);
	}

	/* -------- BOUNCE ANIMATION ---------------*/
	@keyframes bounceDown {
		0% {
			transform: scale(1);
		}
		50% {
			transform: scale(0.85);
		}
		100% {
			transform: scale(1);
		}
	}

	.menu-img.bounce {
		animation: bounceDown 0.25s ease;
	}

	/* -------- Sizing ---------------*/

	/* Start button image sizing */
	.menu-button--start .menu-img {
		height: clamp(4vh, 6rem, 10vh);
		max-width: 90vw;
	}
</style>
