#include <CoreMIDI/MIDIServices.h>
#include <CoreFoundation/CFRunLoop.h>
#include <mach/mach_time.h>

struct client {
    int             wasInit;
    MIDIClientRef   client;
    MIDIPortRef     inputPort;
    MIDIPortRef     outputPort;
    void            *goDriver;
};

struct client* allocateClient();
int init(struct client *c);
void shutdown(struct client *c);
int openInput(struct client *c, MIDIEndpointRef source);
int send(struct client *c, MIDIEndpointRef destination, uint64_t timestamp, uint32_t *words, uint32_t wordCount);
