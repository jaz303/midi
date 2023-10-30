#include <CoreMIDI/MIDIServices.h>
#include <CoreServices/CoreServices.h>
#include <mach/mach_time.h>
#include <assert.h>

#include "_cgo_export.h"

static CFStringRef      kClientName     = CFSTR("jaz303/midi");
static CFStringRef      kInputName      = CFSTR("input");
static CFStringRef      kOutputName     = CFSTR("output");

static_assert(sizeof(void*) >= sizeof(MIDIEndpointRef), "MIDIEndpointRef won't fit in a pointer!");

struct client* allocateClient() {
    void *out = malloc(sizeof(struct client));
    memset(out, 0, sizeof(struct client));
    return out;
}

int init(struct client *c) {
    OSStatus status;

    status = MIDIClientCreate(kClientName, NULL, NULL, &c->client);
    if (status != 0) {
        return status;
    }

    status = MIDIInputPortCreateWithProtocol(c->client, kInputName, kMIDIProtocol_1_0, &c->inputPort, ^void (const MIDIEventList *eventList, void *source) {
        MIDIEventPacket *pkt = (MIDIEventPacket*)&eventList->packet[0];
        for (int i = 0; i < eventList->numPackets; i++) {
            OnReceive(c->goDriver, pkt->timeStamp, source, pkt->words, pkt->wordCount);
            pkt = MIDIEventPacketNext(pkt);
        }
    });
    
    if (status != 0) {
        return status;
    }

    status = MIDIOutputPortCreate(c->client, kOutputName, &c->outputPort);
    if (status != 0) {
        return status;
    }

    return 0;
}

void shutdown(struct client *c) {
    if (c->wasInit) {

    }
    free(c);
}

int openInput(struct client *c, MIDIEndpointRef source) {
    OSStatus status = MIDIPortConnectSource(c->inputPort, source, (void*)((uintptr_t)source));
    return status;
}

int send(struct client *c, MIDIEndpointRef destination, uint64_t timestamp, uint32_t *words, uint32_t wordCount) {
    MIDIEventList lst;
    MIDIEventPacket *pkt = MIDIEventListInit(&lst, kMIDIProtocol_1_0);
    MIDIEventListAdd(&lst, sizeof(MIDIEventList), pkt, timestamp, wordCount, words);
    MIDISendEventList(c->outputPort, destination, &lst);
    return 0;
}
