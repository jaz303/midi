#include <CoreMIDI/MIDIServices.h>
#include <CoreServices/CoreServices.h>
#include <mach/mach_time.h>
#include <assert.h>

#include "_cgo_export.h"

static CFStringRef      kClientName     = CFSTR("jaz303/midi");
static CFStringRef      kInputName      = CFSTR("input");
static CFStringRef      kOutputName     = CFSTR("output");

static MIDIClientRef    client;
static MIDIPortRef      inputPort;
static MIDIPortRef      outputPort;

static_assert(sizeof(void*) >= sizeof(MIDIEndpointRef), "MIDIEndpointRef won't fit in a pointer!");

int init() {
    OSStatus status;

    status = MIDIClientCreate(kClientName, NULL, NULL, &client);
    if (status != 0) {
        return status;
    }

    status = MIDIInputPortCreateWithProtocol(client, kInputName, kMIDIProtocol_1_0, &inputPort, ^void (const MIDIEventList *eventList, void *source) {
        MIDIEventPacket *pkt = &eventList->packet[0];
        for (int i = 0; i < eventList->numPackets; i++) {
            OnReceive(pkt->timeStamp, source, pkt->words, pkt->wordCount);
            pkt = MIDIEventPacketNext(pkt);
        }
    });
    
    if (status != 0) {
        return status;
    }

    status = MIDIOutputPortCreate(client, kOutputName, &outputPort);
    if (status != 0) {
        return status;
    }

    return 0;
}

int openInput(MIDIEndpointRef source, void *sourceAsPointer) {
    OSStatus status = MIDIPortConnectSource(inputPort, source, sourceAsPointer);
    return status;
}

int send(MIDIEndpointRef destination, uint64_t timestamp, uint32_t *words, uint32_t wordCount) {
    MIDIEventList lst;
    MIDIEventPacket *pkt = MIDIEventListInit(&lst, kMIDIProtocol_1_0);
    MIDIEventListAdd(&lst, sizeof(MIDIEventList), pkt, timestamp, wordCount, words);
    MIDISendEventList(outputPort, destination, &lst);
    return 0;
}
