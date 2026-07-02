# DaisyUI Components

Ready-to-copy DaisyUI components. All use Tailwind + DaisyUI classes.

> 📖 For the latest DaisyUI component docs, fetch `https://daisyui.com/llms.txt` —
> it contains current component names, modifiers, themes, and usage rules.

## RULES
Follow this instructions for Daisyui:
- Use daisyUI components properly (btn, card, drawer, modal, navbar, etc.).
- Prefer semantic daisyUI classes over long utility chains.
- Follow daisyUI best practices and theming system.
- Generate clean, responsive, accessible code.


## Modal with Datastar

> ⚠️ **Do NOT use `data-show`** on `<dialog>` elements — it only toggles `display: none`, which doesn't work with DaisyUI modals. DaisyUI requires the `modal-open` CSS class.

**✅ CORRECT: Use `data-class` with `modal-open`**
```html
<dialog class="modal" data-class="{'modal-open': $showDialog}">
  <div class="modal-box">
    <h3 class="font-bold text-lg">Title</h3>
    <div class="modal-action">
      <button data-on:click="$showDialog = false">Close</button>
    </div>
  </div>
</dialog>

<button class="btn btn-primary" data-on:click="$showDialog = true">Open Modal</button>
```

**❌ WRONG: `data-show` on dialog**
```html
<dialog class="modal" data-show="$showDialog">
  <!-- Modal won't display properly -->
</dialog>
```
