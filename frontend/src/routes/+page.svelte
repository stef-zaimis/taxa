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

	<div class="menu-content">
		<img src="/mm/title.png" alt="TAXA Title" class="w-[90vw] max-w-4xl" />

		<button on:click={startQuiz} type="button">
			<img src="/mm/start_button.png" alt="Start Button" class="w-40 md:w-60 cursor-pointer hover:scale-105 transition-transform"/>
		</button>

		<div class="flex gap-6 pt-4 flex-wrap">

			<button on:click={goToOptions} type="button">
				<img src="/mm/options_button.png" alt="Options" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform" on:click={goToOptions} />
			</button>

			<button on:click={goToSettings} type="button">
				<img src="/mm/settings_button.png" alt="Settings" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform"/>
			</button>

			<button on:click={exitGame} type="button">
				<img src="/mm/exit_game_button.png" alt="Exit" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform"/>
			</button>
		</div>
	</div>
</div>

<style>
	.main-menu-wrapper {
		width: 100vw;
		height: 100vh;
		overflow: hidden;
		background-image: url('/mm/bg.png');
		background-size: cover;
		background-position: center;
		background-attachment: fixed;
		position: relative;
	}

	.menu-content {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 2rem;
		padding-left: 1rem;
		padding-right: 1rem;
	}
</style>
