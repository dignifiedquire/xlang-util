#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * Helper to explicitly transfer a slice of bytes across FFI bounds.
 */
typedef struct Message Message;

typedef struct Token {
  struct Slot *slot;
  uint64_t stamp;
} Token;

/**
 * A slot in a channel.
 */
typedef struct Slot {
  /**
   * The current stamp.
   */
  uint64_t stamp;
  /**
   * The message in this slot.
   */
  struct Message *msg;
} Slot;

typedef struct Channel {
  /**
   * The head of the channel.
   *
   * This value is a "stamp" consisting of an index into the buffer, a mark bit, and a lap, but
   * packed into a single `usize`. The lower bits represent the index, while the upper bits
   * represent the lap. The mark bit in the head is always zero.
   *
   * Messages are popped from the head of the channel.
   */
  uint64_t head;
  /**
   * The tail of the channel.
   *
   * This value is a "stamp" consisting of an index into the buffer, a mark bit, and a lap, but
   * packed into a single `usize`. The lower bits represent the index, while the upper bits
   * represent the lap. The mark bit indicates that the channel is disconnected.
   *
   * Messages are pushed into the tail of the channel.
   */
  uint64_t tail;
  /**
   * The buffer holding slots.
   */
  struct Slot *buffer;
  /**
   * The channel capacity.
   */
  uintptr_t cap;
  /**
   * A stamp with the value of `{ lap: 1, mark: 0, index: 0 }`.
   */
  uint64_t one_lap;
  /**
   * If this bit is set in the tail, that means the channel is disconnected.
   */
  uint64_t mark_bit;
} Channel;

struct Channel *new_channel(uint32_t cap);

void drop_channel(struct Channel *channel_ptr);

uint8_t *new_message_bytes(const uint8_t *ptr, uint64_t len);

void drop_message(struct Message *msg);
