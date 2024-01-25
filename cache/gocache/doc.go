package gocache

/***
gocache:
wrapper github.com/patrickmn/go-cache to implement Cache[Value any] interface.
It should be noted that gocache cannot set the memory elimination algorithm and can only clear expired objects by setting cleanupInterval.
It will continue to expand the map until the program crashes.
*/
