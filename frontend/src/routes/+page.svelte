<script lang="ts">
	import { goto } from '$app/navigation';

	function startQuiz() {
		goto('/quiz');
	}

	function goToOptions() {
		//TODO
	}

	function goToSettings() {
		//TODO
	}

	function exitGame() {
		//TODO
	}
</script>

<div class="main-menu-wrapper">
	<div class="menu-wrapper-inner">
		<div class="menu-content">
			<img src="/mm/title.png" alt="TAXA Title" class="w-[90vw] max-w-4xl" />

			<button on:click={startQuiz} type="button">
				<img src="/mm/start_button.png" alt="Start Button" class="menu-button menu-button--start" />
			</button>

			<div class="button-group">

				<button on:click={goToOptions} type="button">
					<img src="/mm/options_button.png" alt="Options" class="menu-button"/>
				</button>

				<button on:click={goToSettings} type="button">
					<img src="/mm/settings_button.png" alt="Settings" class="menu-button"/>
				</button>

				<button on:click={exitGame} type="button">
					<img src="/mm/exit_game_button.png" alt="Exit" class="menu-button"/>
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
		background-image: url('/mm/bg.png');
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
		gap: 2rem;
		width: 100%;
		max-width: 100%;
	}

	.button-group {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 1.5rem;
		padding-top: 1rem;
	}

	@media (min-width: 768px) {
		.button-group {
			gap: 3rem;
		}
	}

/* --------COMMUNAL BUTTONS---------------*/
	.menu-button {
		height: 1.9rem;
		cursor: pointer;
		transition: transform 0.25s ease;
		object-fit: contain;
	}

	.menu-button:hover {
		transform: scale(0.9);
	}

	@media (min-width: 768px) {
		.menu-button {
			height: 3.6rem;
		}
	}

/* ---------------START BUTTON-----------*/
	.menu-button--start {
		height: 4rem;
	}

	@media (min-width: 768px) {
		.menu-button--start {
			height: 5.5rem;
		}
	}

	.menu-button--start:hover {
		transform: scale(1.15);
	}

</style>
