#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct Token {
  struct Slot *slot;
  uint64_t stamp;
} Token;

typedef struct Slot {
  uint64_t stamp;
  uint8_t *msg_ptr;
  uint64_t msg_len;

} Slot;

typedef struct Channel {
  uint64_t head;
  uint64_t tail;
  struct Slot *buffer;
  uintptr_t cap;
  uint64_t one_lap;
  uint64_t mark_bit;
} Channel;

struct Channel *new_channel(uint32_t cap);

void drop_channel(struct Channel *channel_ptr);

uint8_t* channel_send(struct Channel *channel_ptr, uint8_t *ptr, uint64_t len);

uint8_t* channel_recv(struct Channel *channel_ptr, uint64_t *out_len);
uint8_t* channel_try_recv(struct Channel *channel_ptr, uint64_t *out_len);

uint8_t *new_message_bytes(const uint8_t *ptr, uint64_t len);

void drop_message_bytes(uint8_t *ptr, uint64_t len);

uintptr_t slotSize();
