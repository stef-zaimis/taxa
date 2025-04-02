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

<div class="relative w-screen h-screen overflow-hidden bg-fixed bg-center bg-cover" style="background-image: url('/mm/bg.png');">

	<div class="absolute inset-0 flex flex-col items-center justify-center gap-8 px-4">
		<img src="/mm/title.png" alt="TAXA Title" class="w-[90vw] max-w-4xl" />

		<img src="/mm/start_button.png" alt="Start Button" class="w-40 md:w-60 cursor-pointer hover:scale-105 transition-transform" on:click{startQuiz} />

		<div class="flex gap-6 pt-4 flex-wrap: wrap">
			<img src="/mm/options_button.png" alt="Options" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform" on:click={goToOptions} />
			<img src="/mm/settings_button.png" alt="Settings" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform" on:click={goToSettings} />
			<img src="/mm/exit_game_button.png" alt="Exit" class="w-28 md:w-36 cursor-pointer hover:scale-105 transition-transform" on:click={exitGame} />
		</div>
	</div>
</div>
