<script lang="ts">
	import { goto } from '$app/navigation';

	function startQuiz() {
		goto('/quiz');
	}
</script>

<h1>Taxonomic Quiz</h1>
<p>Welcome. Click on the button below to begin:</p>

<button on:click={startQuiz} style="padding: 1rem 2rem; font-size: 1.25rem; margi-top: 1rem;"> Start Quiz </button>
