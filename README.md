# artifactsmmo.com
My implementation of an artifacts-mmo API client, written in Golang

# TODO
- [x] Display inventory items in UI
- [x] Display for equipped items
- [x] Cache bank state to share between kernels
    - [x] a `get-bank` command to manually update the bank state - ended up with `list-bank`
    - [x] a display of bank items, can be toggled into view using something like `ctrl+b` (`<C-b>`) or maybe just `<`, similar to how orders work (then orders could be viewed with `>`) - `list-bank` does the displaying - I did this PER CHARACTER, which was a mistake, but it's actually highly useful as it allows you to compare/contrast lists
    - [x] a `bank-has` command to filter entire bank list for searching - `list-bank [part-of-a-code]` and also `hide-bank`
    - [x] Use the cached bank state to make bank decisions, only busting cache when a bank operation is done
- [x] Something like generators that looks out for events and acts on them while they're active
- [ ] display of all active generators
- [ ] Fighting armor + weapon set picker
    - [ ] `find` command that accepts params and finds items that match that description
- [ ] `move to:<monster/resource>` to move to the closest monster/resource
- [ ] `go-equip` or something to equip from bank
- [ ] Generators do not "lock in" event locations when active - `gen make demon_horn` when demons are present ends up getting stuck after the event ends
- [ ] Revised fight simulator
    - [ ] Includes items
    - [ ] Includes crits
    - [ ] Includes enemy dodging
    - [ ] Can be forced to include nothing
- [ ] Scrolling on logs
- [ ] Command history per kernel
- [ ] Cache static responses
    - [x] events
    - [ ] map locations
- [ ] Improve error handling for generators
- [ ] Better inventory management in generators
- [ ] Season 4 market automation

## Event watcher idea
- goroutine polls for events, kernel state change, almost anything we want, sends updates to string channel owned by kernels
- kernels will check the channel at the beginning of every command-checking loop
- some command exists to set a cause-effect, something like `when some-event-start then gen make demon_horn`, operatives are `when <event-id> then ...<full command to enter>`
- game events have ids like `demon_portal-begin` and `demon_portal-end`, event watcher is stateful, aware of when to send the begin and end commands for events
- need some way to display and then also manipulate when/thens
    - maybe all of this state shouldn't actually be gui-controlled, instead using a state-file read at startup
- may need some way to prioritize commands? what if two events are on-going together?

## Event watcher alternative
- stackable generators
- generator for event that can check for event state and take an action, or pass to the next generator
```
gen level gearcrafting # sets generator 0
gen 0 level fight # overwrites generator 0
gen 1 make demon_horn # sets generator 1
gen level alchemy # sets generator 2

clear-gen # clears all
pause-gen # pauses all

clear-gen 0 # clears 0
pause-gen 1 # pauses 1
```
