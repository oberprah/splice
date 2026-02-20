# Splice

## Motivation

With AI, even more code will be produced than before.
Which means we need to review more code than before.

Right now I'm not super happy with my current workflow.
I'm using tig and delta in the terminal and for more complex diffs the diff view in IntelliJ.
The diff view in IntelliJ I like.
However, the shortcuts, I do not like and I found no way to change them in a way I like them.

I gave vimdiff a try and thought there I have more configuration options and flexibility.
But it turns out that it isn't that easy to configure things the way I would like to have them.

Thus, I thought it might be the easiest to build something from scratch.
Which also might be a good option to use AI tools and probably have a public GitHub repo.

## Features

Most important would be a good diff view for me.

What I don't like with `delta` is that for large diffs there is no overview about all files.
One sees all the changes, but doesn't have an option to see all changed files.
This is solved a lot better in IntelliJ, since there exists a view about all changed files.

However, in IntelliJ it isn't easy to open that view - it takes multiple shortcuts to get there.
Furthermore, I wasn't able to change and set shortcuts in the diff view
And needed to use the mouse for a few actions.

- Shortcut to open a diff view.
- Shortcuts to navigate inside a diff view (between the files, the changes, inside a file).
- Have an overview view of all changed files.

Furthermore additional features like git logs would be nice to include.

[guti](https://github.com/altsem/gitu) is also a nice tool which goes in the same direction.
But I do not like the diff view there.
It doesn't support side-by-side view

## Technology

I'm a terminal guy.
I heavily use Claude code and vim lately.
Thus, it should become a TUI app.

Recently there was an [article about Claude code]( https://newsletter.pragmaticengineer.com/p/how-claude-code-is-built), which is as well a terminal app.
They decided to use [ink](https://github.com/vadimdemedes/ink).

> Claude Code’s tech stack:
>
> - TypeScript: Claude Code is built on this language
> - React with Ink: the UI is written in React, using the Ink framework for interactive
>   command-line elements
> - Yoga: the layout system, open sourced by Meta. It's a constraints-based layout that
>   works nicely. Terminal-based applications have the disadvantage of needing to support all
>   sizes of terminals, so you need a layout system to do this pragmatically
> - Bun: for building and packaging. The team chose it for speed compared to other build
>   systems like Webpack, Vite, and others.

> The tech stack was chosen to be "on distribution" for the Claude model. In AI, there are
> the terms "on distribution" and "off distribution." "On distribution" means the model
> already knows how to do it, and "off distribution" means it's not good at it.
>
> The team wanted an "on distribution" tech stack for Claude that it was already good at.
> TypeScript and React are two technologies the model is very capable with, so were a logical
> choice. However, if the team had chosen a more exotic stack Claude isn't that great with,
> then it would be an "off distribution" stack. Boris sums it up:
>
> "With an off-distribution stack, the model can still learn it. But you have to show it the
> ropes and put in the work. We wanted a tech stack which we didn't need to teach: one where
> Claude Code could build itself. And it's working great; around 90% of Claude Code is written
> with Claude Code".

I think this tech stack would also be a good fit for my app.
Especially Typescript, React and Ink.

Other options would be [bubbletea](https://github.com/charmbracelet/bubbletea) which is a TUI framework written in Go.
Or [ratatui](https://github.com/ratatui/ratatui) which is a TUI framework written in Rust.

Both languages are unfamiliar for me.
Furthermore, TypeScript is also more famous and thus AI will be better at it.
