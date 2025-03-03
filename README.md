# artifactsmmo.com
My implementation of an artifacts-mmo API client, written in Golang

# TODO
- [x] Display inventory items in UI
- [ ] Display for equipped items
- [ ] Cache bank state to share between kernels
    - [ ] a `get-bank` command to manually update the bank state
    - [ ] a display of bank items, can be toggled into view using something like `ctrl+b` (`<C-b>`) or maybe just `<`, similar to how orders work (then orders could be viewed with `>`)
    - [ ] a `bank-has` command to filter entire bank list for searching
- [ ] Something like generators that looks out for events and acts on them while they're active
- [ ] Cache static responses
- [ ] Revised fight simulator
    - [ ] Includes items
    - [ ] Includes crits
    - [ ] Includes enemy dodging
    - [ ] Can be forced to include nothing
- [ ] Fighting armor + weapon set picker
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
