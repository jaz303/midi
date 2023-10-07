#include <CoreMIDI/MIDIServices.h>
#include <mach/mach_time.h>

int init();
int openInput(MIDIEndpointRef source, void *sourceAsPointer);
int send(MIDIEndpointRef destination, uint64_t timestamp, uint32_t *words, uint32_t wordCount);
