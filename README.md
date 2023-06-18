# Matrix-FeedReader

A very simple bot that posts new entries from a Miniflux RSS reader to a Matrix room.

Miniflux already has a Matrix integration and can post the entries itself, but this bot adds two things:

* After posting it marks the entry as read in Miniflux.
* It posts every entry as a separate message, which makes it easier to create other bots that can interact with them.
