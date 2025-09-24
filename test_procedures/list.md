# List Component Test Procedure

## Basic Navigation Features

### 1. Single Item Selection
**Test Steps:**
- Launch list with multiple items
- Use Up/Down arrows or controller D-pad to navigate
- Verify selected item is highlighted
- Press A button or Enter to select
- Verify component exits with selected item

**Expected Results:**
- Navigation should move selection indicator
- Selected item should be visually highlighted
- Selection should wrap around (last item → first item, first item → last item)

### 2. Page Navigation (Left/Right)
**Test Steps:**
- Launch list with more items than `MaxVisibleItems`
- Press Left/Right arrows or controller Left/Right
- Verify page jumps by `MaxVisibleItems` count
- Test at the beginning and end of the list

**Expected Results:**
- Left: Jump to the previous page or beginning
- Right: Jump to the next page or end
- Should handle edge cases properly

### 3. Scrolling Viewport
**Test Steps:**
- Create a list with more than nine items (default `MaxVisibleItems`)
- Navigate through items
- Verify viewport scrolls to keep the selected item visible

**Expected Results:**
- Only `MaxVisibleItems` should be visible at once
- Viewport should scroll smoothly to keep the selection in view

---

## Multi-Select Mode Features

### 4. Toggle Multi-Select Mode
**Test Steps:**
- Launch list with `EnableMultiSelect: true`
- Press Space key or Select button (configurable via `MultiSelectKey`/`MultiSelectButton`)
- Verify mode indicator appears
- Press again to exit multi-select mode

**Expected Results:**
- Visual indicator should show multi-select mode is active
- Checkboxes (☐/☑) should appear next to items
- Exiting should clear all selections

### 5. Multi-Select Item Toggling
**Test Steps:**
- Enter multi-select mode
- Navigate to items and press A button
- Verify items get selected/deselected
- Test with items marked `NotMultiSelectable`

**Expected Results:**
- A button should toggle selection (☐ ↔ ☑)
- `NotMultiSelectable` items should not be toggleable
- Multiple items can be selected simultaneously

### 6. Multi-Select Confirmation
**Test Steps:**
- Select multiple items in multi-select mode
- Press Enter or Start button
- Verify component exits with all selected items

**Expected Results:**
- Component should return with all selected items
- Should work even if no items are selected

---

## Reorder Mode Features

### 7. Toggle Reorder Mode
**Test Steps:**
- Launch list with `EnableReordering: true`
- Press Space key or Select button (configurable via `ReorderKey`/`ReorderButton`)
- Verify reorder indicator appears
- Exit by pressing non-directional key

**Expected Results:**
- Selected item should show reorder indicator (↕)
- Mode should exit on non-directional input

### 8. Item Reordering
**Test Steps:**
- Enter reorder mode
- Use Up/Down to move items one position
- Use Left/Right to move items by page amount
- Verify items swap positions correctly

**Expected Results:**
- Items should physically change positions in the list
- Selection should move with the reordered item
- Should respect list boundaries (no moving beyond first/last)

---

## Text Handling Features

### 9. Text Scrolling (Long Text)
**Test Steps:**
- Create items with text longer than display width
- Navigate to these items
- Wait for automatic text scrolling to begin
- Verify scrolling direction changes and pause behavior

**Expected Results:**
- Long text should scroll horizontally when focused
- Should pause at direction changes
- Scroll speed should match `ScrollSpeed` setting

### 10. Text Truncation (Non-focused)
**Test Steps:**
- Create items with long text
- Navigate away from them
- Verify text is truncated with "..." when not focused

**Expected Results:**
- Non-focused long text should be truncated
- Should show ellipsis indicator

### 11. Title Scrolling
**Test Steps:**
- Set a very long title
- Launch list
- Verify title scrolls if it exceeds screen width

**Expected Results:**
- Long titles should scroll horizontally
- Should follow same scrolling behavior as item text

---

## Image Display Features

### 12. Selected Item Images
**Test Steps:**
- Launch list with `EnableImages: true`
- Create items with `ImageFilename` set
- Navigate between items with and without images
- Verify images display on the right side

**Expected Results:**
- Images should appear when item is selected
- Should scale to appropriate size
- Should handle missing image files gracefully

---

## Help System Features

### 13. Help Overlay
**Test Steps:**
- Launch list with `EnableHelp: true` and `HelpText` provided
- Press H key or Menu button
- Navigate help content with Up/Down
- Exit help with any other key

**Expected Results:**
- Help overlay should appear over list
- Should be scrollable if content is long
- Should exit cleanly and return to list

---

## Input Handling Features

### 14. Keyboard vs Controller Input
**Test Steps:**
- Test all features using keyboard keys
- Test same features using controller buttons
- Verify both input methods work identically

**Expected Results:**
- Both input methods should provide identical functionality
- No conflicts or unexpected behavior

### 15. Held Button Repeat
**Test Steps:**
- Hold down directional buttons (Up/Down/Left/Right)
- Verify initial delay before repeat starts
- Verify repeat rate after initial delay

**Expected Results:**
- Should have initial delay (`repeatDelay`)
- Should repeat at consistent interval (`repeatInterval`)
- Should work in both normal and reorder modes

### 16. Input Delay
**Test Steps:**
- Navigate rapidly through items
- Verify navigation respects `InputDelay` setting
- Test with very short delay values

**Expected Results:**
- Should prevent input flooding
- Navigation should feel responsive but controlled

---

## Visual Features

### 17. Empty List Handling
**Test Steps:**
- Launch list with no items
- Verify empty message displays
- Test that only exit actions work (B button, help)

**Expected Results:**
- Should display `EmptyMessage`
- Should center message on screen
- Should handle multi-line empty messages

### 18. Footer Display
**Test Steps:**
- Launch list with `FooterHelpItems` configured
- Verify footer appears at bottom with help text
- Test with different footer configurations

**Expected Results:**
- Footer should appear at screen bottom
- Should show appropriate button hints
- Should respect margins

### 19. Theme Integration
**Test Steps:**
- Launch list and verify colors match theme
- Test focused vs non-focused item colors
- Verify background colors and text colors

**Expected Results:**
- Should use theme colors consistently
- Focused items should have distinct appearance
- Should be visually clear and readable

---

## Configuration Features

### 20. Custom Button Mapping
**Test Steps:**
- Configure custom `MultiSelectKey`/`MultiSelectButton`
- Configure custom `ReorderKey`/`ReorderButton`
- Verify custom mappings work correctly

**Expected Results:**
- Custom key mappings should work as configured
- Should not conflict with other functionality

### 21. Disable Back Button
**Test Steps:**
- Launch list with `DisableBackButton: true`
- Press B button or equivalent
- Verify exit is prevented

**Expected Results:**
- B button should not exit the component
- Should remain in list interface

### 22. Action Button
**Test Steps:**
- Launch list with `EnableAction: true`
- Press X button
- Verify component exits with `ActionTriggered: true`

**Expected Results:**
- X button should trigger action and exit
- Should set appropriate return value

---

## Callback Features

### 23. OnSelect Callback
**Test Steps:**
- Configure `OnSelect` callback
- Select items in both single and multi-select modes
- Verify callback is called with correct parameters

**Expected Results:**
- Callback should be called on item selection
- Should receive correct index and item reference

### 24. OnReorder Callback
**Test Steps:**
- Configure `OnReorder` callback
- Perform item reordering
- Verify callback receives correct from/to indices

**Expected Results:**
- Callback should be called when items are reordered
- Should receive accurate position information

---

## Edge Cases and Error Handling

### 25. Invalid Initial Selection
**Test Steps:**
- Set `SelectedIndex` to invalid value (negative or beyond list length)
- Launch list
- Verify it defaults to valid selection

**Expected Results:**
- Should default to index 0 for invalid initial selections
- Should not crash or show errors

### 26. Rapid Input Handling
**Test Steps:**
- Perform very rapid button presses
- Hold multiple directional buttons simultaneously
- Verify stable behavior

**Expected Results:**
- Should handle rapid input gracefully
- Should not queue excessive actions
- Should maintain a stable state