# Tempo Demo Script

A step-by-step guide for recording a demo of Tempo's features.

## Prerequisites

Before recording:
1. Have a local Temporal server running with sample workflows
2. Create a few workflows in different states (Running, Completed, Failed)
3. Optionally set up a second profile to demo profile switching
4. Terminal size: recommend 120x35 or larger for good visibility

## Demo Sequence

### Scene 1: Application Launch & Namespace View (10s)

**Goal:** Show the clean startup and namespace browsing

1. Launch Tempo: `tempo`
2. Pause briefly on the namespace list
3. Show the status bar at the bottom (profile name, connection status)
4. Navigate with `j`/`k` through namespaces
5. Press `Enter` to select the `default` namespace

### Scene 2: Workflow List Overview (15s)

**Goal:** Showcase the multi-panel layout and workflow browsing

1. Arrive at the workflow list view
2. Point out the layout:
   - Left panel: workflow list with ID, type, status, times
   - Right panel: preview with workflow details
3. Navigate through workflows with `j`/`k`
4. Show how the preview updates automatically
5. Point out the status bar counts (Running/Completed/Failed)

### Scene 3: Filtering & Search (20s)

**Goal:** Demonstrate the powerful filtering capabilities

1. Press `/` to open local filter
2. Type a partial workflow ID or type to filter
3. Clear with `Esc`
4. Press `F` to open the visibility query modal
5. Show example query: `WorkflowType = "OrderWorkflow"`
6. Execute the query and show filtered results
7. Press `C` to clear the query
8. (Optional) Press `f` to show query templates

### Scene 4: Workflow Detail View (15s)

**Goal:** Show detailed workflow inspection

1. Press `Enter` on a workflow to open detail view
2. Show the workflow metadata section
3. Press `i` to show input/output modal
4. Close modal with `Esc`
5. Navigate through the tabs/sections

### Scene 5: Event History Views (20s)

**Goal:** Demonstrate the three event visualization modes

1. Press `e` to enter the event history view
2. **List View:** Show the default chronological event list
3. Navigate events with `j`/`k`
4. Press `d` on an event to show the detail modal (JSON display)
5. Close modal, then press `Tab` to switch to **Tree View**
6. Show the hierarchical event relationships
7. Press `Tab` again for **Timeline View**
8. Show the chronological timeline visualization
9. Press `Esc` to return to workflow detail

### Scene 6: Workflow Actions (15s)

**Goal:** Show workflow management capabilities

1. On a **Running** workflow:
   - Press `s` to open Signal modal (show the form, then cancel)
   - Press `Q` to open Query modal (show the form, then cancel)
   - Press `c` to show Cancel confirmation (show warning, then cancel)
2. On a **Completed/Failed** workflow:
   - Press `R` to show Reset options
   - Cancel to exit

### Scene 7: Batch Operations (15s)

**Goal:** Demonstrate multi-select and batch actions

1. Return to workflow list
2. Press `v` to enter selection mode
3. Use `Space` to select multiple workflows
4. Show selected count updating
5. Press `c` to show batch cancel confirmation (shows count breakdown)
6. Cancel the operation
7. Press `v` to exit selection mode

### Scene 8: Theme Switching (10s)

**Goal:** Showcase the theme system

1. Press `T` to open theme selector
2. Navigate through themes with `j`/`k`
3. Show live preview updating as you navigate
4. Select a different theme (e.g., switch from dark to light or vice versa)
5. Show the UI with the new theme

### Scene 9: Profile Management (10s)

**Goal:** Show connection profile features

1. Press `P` to open profile selector
2. Show the list of available profiles
3. (If multiple profiles exist) Switch to a different profile
4. Show the status bar updating with new profile name

### Scene 10: Help System (5s)

**Goal:** Show context-aware help

1. Press `?` to open help modal
2. Show the keybindings for the current view
3. Close with `Esc`

### Scene 11: Additional Features (Optional, 15s)

**Task Queues:**
1. From workflow list, press `t` to view task queues
2. Browse the task queue list
3. Press `Esc` to return

**Schedules:**
1. From workflow list, press `s` (capital S from namespace) to view schedules
2. Browse schedules
3. Press `Esc` to return

**Auto-Refresh:**
1. Press `a` to enable auto-refresh
2. Show the indicator that auto-refresh is active
3. Press `a` again to disable

**Copy Workflow ID:**
1. On a workflow, press `y` to copy ID to clipboard
2. Show the confirmation feedback in preview

---

## Recording Tips

1. **Pacing:** Move slowly enough for viewers to follow, pause 1-2s after each action
2. **Terminal size:** Use a large terminal for readability
3. **Font size:** Consider increasing terminal font for better visibility in compressed GIFs
4. **Clean state:** Start with a clean terminal, no other notifications
5. **Sample data:** Use realistic-looking workflow IDs and types
6. **Theme:** Start with a visually appealing dark theme (tokyonight-night or catppuccin-mocha recommended)

## Suggested Sample Workflows

Before recording, create these workflows for variety:

```bash
# Running workflow
temporal workflow start --type LongRunningProcess --task-queue demo

# Completed workflow
temporal workflow start --type QuickTask --task-queue demo

# Failed workflow (if possible)
temporal workflow start --type FailingWorkflow --task-queue demo
```

## Quick Demo (60 seconds)

For a shorter demo, focus on these core features:

1. Launch and navigate to workflow list (5s)
2. Browse workflows, show preview panel (10s)
3. Use filter `/` to search (10s)
4. Enter workflow detail, press `i` for I/O (10s)
5. View event history, switch between views (15s)
6. Switch themes with `T` (5s)
7. Show help with `?` (5s)

## Feature Highlight Clips

Consider creating separate short clips for specific features:

- **filtering.gif** - Just the search/filter functionality
- **events.gif** - Event history views (list, tree, timeline)
- **themes.gif** - Theme switching showcase
- **batch.gif** - Multi-select and batch operations
- **actions.gif** - Workflow actions (signal, query, cancel)
