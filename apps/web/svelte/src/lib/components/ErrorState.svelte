<script lang="ts">
  import { AlertTriangle, RotateCcw, ClipboardCopy, Check } from 'lucide-svelte';

  interface Action {
    label: string;
    href?: string;
    onClick?: () => void;
    primary?: boolean;
  }

  interface Props {
    title?: string;
    message?: string;
    actions?: Action[];
    diagnosticInfo?: string;
    minHeight?: string;
  }

  let {
    title = "Something went wrong",
    message = "Make sure your Xuva server is running, then try again.",
    actions = [],
    diagnosticInfo = '',
    minHeight = '60vh',
  }: Props = $props();

  let copied = $state(false);

  async function copyDiagnostics() {
    if (!diagnosticInfo) return;
    try {
      await navigator.clipboard.writeText(diagnosticInfo);
      copied = true;
      setTimeout(() => { copied = false; }, 2000);
    } catch {
      // Clipboard unavailable — silently ignore
    }
  }
</script>

<div
  class="flex flex-col items-center justify-center gap-3 px-6 text-center"
  style="min-height: {minHeight}"
>
  <div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-foreground/[0.06] text-muted-foreground">
    <AlertTriangle class="h-5 w-5" />
  </div>

  <div class="space-y-1">
    <p class="text-base font-medium text-foreground/80">{title}</p>
    <p class="max-w-xs text-sm text-muted-foreground">{message}</p>
  </div>

  {#if actions.length > 0}
    <div class="mt-2 flex flex-wrap items-center justify-center gap-2">
      {#each actions as action}
        {#if action.href}
          <a
            href={action.href}
            class="hairline inline-flex items-center gap-1.5 rounded-full px-5 py-2.5 text-sm font-medium transition-colors
              {action.primary
                ? 'bg-foreground text-background hover:bg-foreground/90'
                : 'bg-foreground/[0.06] text-muted-foreground hover:text-foreground'}"
          >
            {action.label}
          </a>
        {:else}
          <button
            type="button"
            onclick={action.onClick}
            class="hairline inline-flex items-center gap-1.5 rounded-full px-5 py-2.5 text-sm font-medium transition-colors
              {action.primary
                ? 'bg-foreground text-background hover:bg-foreground/90'
                : 'bg-foreground/[0.06] text-muted-foreground hover:text-foreground'}"
          >
            <RotateCcw class="h-3.5 w-3.5" />
            {action.label}
          </button>
        {/if}
      {/each}
    </div>
  {/if}

  {#if diagnosticInfo}
    <button
      type="button"
      onclick={copyDiagnostics}
      class="mt-1 inline-flex items-center gap-1.5 text-xs text-muted-foreground/60 transition-colors hover:text-muted-foreground"
      title="Copy diagnostic information to clipboard"
    >
      {#if copied}
        <Check class="h-3 w-3" />
        Copied
      {:else}
        <ClipboardCopy class="h-3 w-3" />
        Copy diagnostic info
      {/if}
    </button>
  {/if}
</div>
