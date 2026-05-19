<script lang="ts">
  import { Delete } from 'lucide-svelte';

  interface Props {
    title?: string;
    subtitle?: string;
    error?: string;
    loading?: boolean;
    onsubmit: (pin: string) => void;
    oncancel?: () => void;
  }

  let {
    title = 'Enter PIN',
    subtitle = '',
    error = '',
    loading = false,
    onsubmit,
    oncancel,
  }: Props = $props();

  let digits = $state<string[]>([]);

  const PIN_LENGTH = 4;

  function press(digit: string) {
    if (digits.length < PIN_LENGTH && !loading) {
      digits = [...digits, digit];
      if (digits.length === PIN_LENGTH) {
        onsubmit(digits.join(''));
      }
    }
  }

  function backspace() {
    if (digits.length > 0 && !loading) {
      digits = digits.slice(0, -1);
    }
  }

  /** Allow keyboard entry. */
  function handleKey(e: KeyboardEvent) {
    if (e.key >= '0' && e.key <= '9') press(e.key);
    else if (e.key === 'Backspace') backspace();
    else if (e.key === 'Escape') oncancel?.();
  }

  /** Reset digits so the parent can ask for another attempt. */
  export function reset() {
    digits = [];
  }

  // Auto-reset on error change so user can retry.
  $effect(() => {
    if (error) digits = [];
  });
</script>

<svelte:window onkeydown={handleKey} />

<!-- Modal backdrop -->
<div
  class="fixed inset-0 z-[200] flex items-center justify-center bg-black/70 backdrop-blur-sm"
  onclick={(e) => { if (e.target === e.currentTarget) oncancel?.(); }}
  onkeydown={(e) => { if (e.key === 'Escape') oncancel?.(); }}
  role="presentation"
  tabindex="-1"
>
  <div class="w-full max-w-xs rounded-3xl border border-border bg-surface-elevated p-8 shadow-2xl">
    <!-- Title -->
    <div class="mb-6 text-center">
      <h2 class="text-xl font-semibold text-foreground">{title}</h2>
      {#if subtitle}
        <p class="mt-1 text-sm text-muted-foreground">{subtitle}</p>
      {/if}
    </div>

    <!-- PIN dots -->
    <div class="mb-6 flex justify-center gap-4">
      {#each Array(PIN_LENGTH) as _, i (i)}
        <div
          class={`h-4 w-4 rounded-full border-2 transition-all duration-150 ${
            i < digits.length
              ? 'border-primary bg-primary scale-110'
              : 'border-border bg-transparent'
          }`}
        ></div>
      {/each}
    </div>

    <!-- Error message -->
    {#if error}
      <p class="mb-4 text-center text-sm font-medium text-red-400">{error}</p>
    {/if}

    <!-- Number grid -->
    <div class="grid grid-cols-3 gap-3">
      {#each ['1','2','3','4','5','6','7','8','9'] as digit (digit)}
        <button
          type="button"
          onclick={() => press(digit)}
          disabled={loading || digits.length >= PIN_LENGTH}
          class="flex h-14 items-center justify-center rounded-2xl border border-border bg-surface text-lg font-semibold text-foreground transition-all hover:bg-surface-elevated hover:scale-105 active:scale-95 disabled:opacity-40"
        >
          {digit}
        </button>
      {/each}

      <!-- bottom row: cancel · 0 · backspace -->
      {#if oncancel}
        <button
          type="button"
          onclick={oncancel}
          class="flex h-14 items-center justify-center rounded-2xl text-sm text-muted-foreground transition-colors hover:text-foreground"
        >
          Cancel
        </button>
      {:else}
        <div></div>
      {/if}

      <button
        type="button"
        onclick={() => press('0')}
        disabled={loading || digits.length >= PIN_LENGTH}
        class="flex h-14 items-center justify-center rounded-2xl border border-border bg-surface text-lg font-semibold text-foreground transition-all hover:bg-surface-elevated hover:scale-105 active:scale-95 disabled:opacity-40"
      >
        0
      </button>

      <button
        type="button"
        onclick={backspace}
        disabled={loading || digits.length === 0}
        class="flex h-14 items-center justify-center rounded-2xl text-muted-foreground transition-colors hover:text-foreground disabled:opacity-30"
        aria-label="Backspace"
      >
        <Delete class="h-5 w-5" />
      </button>
    </div>

    {#if loading}
      <div class="mt-5 flex justify-center">
        <div class="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
      </div>
    {/if}
  </div>
</div>
