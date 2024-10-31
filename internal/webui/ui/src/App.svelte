<script lang="ts">
  import FileButton from "./lib/FileButton.svelte";
  import { active, watcher } from "./lib/watcher.svelte";
</script>

<div id="wrapper">
  <section id="sidebar">
    <input type="search" placeholder="Search files" />

    {#each watcher.files as file}
      <FileButton
        {file}
        onclick={(name) => {
          active.file = watcher.files.find((f) => f.name == name);
        }}
      />
    {/each}
  </section>

  <section id="detatils">
    <input type="search" placeholder="Search" />
    {#if !active.file}
      <p>Select a file</p>
    {:else}
      {#each active.file.subTests as subTest}
        <FileButton
          file={subTest}
          onclick={(name) => {
            active.test = active.file?.subTests.find((f) => f.name == name);
          }}
        />
      {/each}
    {/if}
  </section>

  <main>
    {#if active.test}
      {#each active.test.errors || [] as { message }}
        <pre>{message}</pre>
      {:else}
        <p>No errors</p>
      {/each}
    {:else}
      <p>Select a test</p>
    {/if}
  </main>
</div>

<style>
  #wrapper {
    display: grid;
    grid-template-columns: 300px 300px 1fr;
    gap: 1rem;
  }

  #sidebar {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  input {
    width: 100%;
  }

  pre {
    tab-size: 4;
  }
</style>
