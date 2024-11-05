<script lang="ts">
  let { message }: { message: string } = $props();

  const lines = $derived(
    message.split("\n").map((line, i) => ({
      line: i,
      m: line,
      add: line.startsWith("+"),
      del: line.startsWith("-"),
    }))
  );

  const maxLines = lines.length;

  const padZero = (num: number) =>
    num.toString().padStart(maxLines.toString().length, "0");
</script>

<pre>{#each lines as { m, add, del, line }}<span
      class:add
      class:del
      data-line={padZero(line)}
      >{m}
</span>{/each}</pre>

<style>
  pre {
    white-space: pre-wrap;
    font-family: monospace;
    tab-size: 4;
  }

  .add {
    color: rgb(145, 233, 145);
  }

  .del {
    color: rgb(196, 115, 115);
  }

  span[data-line] {
    display: block;

    &:hover {
      background-color: rgba(153, 151, 151, 0.1);
    }

    &::before {
      content: attr(data-line);
      display: inline-block;
      margin-right: 1em;
      color: #666;
    }
  }
</style>
