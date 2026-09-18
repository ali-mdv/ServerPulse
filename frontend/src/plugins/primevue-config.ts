import type { PrimeVueConfiguration } from "primevue/config";

export const primeVueConfig: PrimeVueConfiguration = {
  ripple: false,
  pt: {
    dialog: {
      root: {
        class: "rounded-lg overflow-hidden border border-border shadow-xl bg-card text-card-foreground",
      },
      header: {
        class:
          "flex items-center justify-between px-5 py-3 border-b border-border bg-card text-card-foreground",
      },
      headerTitle: {
        class: "font-semibold text-base",
      },
      headerIcons: {
        class: "flex items-center gap-1",
      },
      content: {
        class: "p-5 bg-card text-card-foreground",
      },
      mask: {
        class: "bg-black/50",
      },
      closeButton: {
        class:
          "btn btn-ghost btn-icon text-muted-foreground hover:text-foreground",
      },
      closeIcon: {
        class: "w-4 h-4",
      },
    },
    toast: {
      root: {
        class: "z-[60]",
      },
      message: ({ state }) => ({
        class: [
          "rounded-lg border bg-popover text-popover-foreground shadow-lg overflow-hidden mb-2",
          state.messages?.length && [
            state.messages[0].severity === "success" &&
              "border-l-4 border-l-success",
            state.messages[0].severity === "info" &&
              "border-l-4 border-l-info",
            state.messages[0].severity === "warn" &&
              "border-l-4 border-l-warning",
            state.messages[0].severity === "error" &&
              "border-l-4 border-l-critical",
          ],
        ],
      }),
      messageContent: {
        class: "flex items-start gap-3 p-3",
      },
      messageText: {
        class: "flex-1",
      },
      summary: {
        class: "font-semibold text-sm text-foreground",
      },
      detail: {
        class: "text-sm text-muted-foreground mt-0.5",
      },
      messageIcon: ({ state }) => ({
        class: [
          "w-5 h-5 mt-0.5 shrink-0",
          state.messages?.length && [
            state.messages[0].severity === "success" && "text-success",
            state.messages[0].severity === "info" && "text-info",
            state.messages[0].severity === "warn" && "text-warning",
            state.messages[0].severity === "error" && "text-critical",
          ],
        ],
      }),
      closeButton: {
        class: "btn btn-ghost btn-icon text-muted-foreground",
      },
      closeIcon: {
        class: "w-4 h-4",
      },
    },
  },
};
