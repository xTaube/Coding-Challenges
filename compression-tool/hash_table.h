#include <stdint.h>

struct UTF8CharRow {
  wchar_t wc;
  uint64_t count;
};

struct UTF8CharTable {
  struct UTF8CharRow **rows;
  uint32_t capacity;
  uint32_t lenght;
};
