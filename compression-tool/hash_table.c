#include <stdint.h>

#define FNV_OFFSET 14695981039346656037UL
#define FNV_PRIME 1099511628211UL

uint64_t hash_wc(const wchar_t wc) {
  uint64_t hash = FNV_OFFSET;
  hash ^= (uint64_t)wc;
  hash *= FNV_PRIME;
  return hash;
}

void new_entry(struct UTF8CharTable *table, wchar_t wc, uint32_t index) {
  struct UTF8CharRow *row =
      (struct UTF8CharRow *)malloc(sizeof(struct UTF8CharRow));
  row->wc = wc;
  row->count = 1;

  table->rows[index] = row;
  table->lenght++;
}

void increment_counter(struct UTF8CharTable *table, wchar_t wc) {
  uint32_t index = hash_wc(wc) % table->capacity;

  while (table->rows[index] != NULL && table->rows[index]->wc != wc) {
    index = (index + 1) % table->capacity;
  }

  if (table->rows[index] == NULL) {
    new_entry(table, wc, index);
  } else {
    table->rows[index]->count++;
  }
}

uint32_t get_count(struct UTF8CharTable *table, wchar_t wc) {
  uint32_t index = hash_wc(wc) % table->capacity;

  while (table->rows[index]->wc != wc) {
    index = (index + 1) % table->capacity;
  }

  return table->rows[index]->count;
}

void free_utf8_char_table(struct UTF8CharTable *table) {
  for (uint32_t i = 0; i < table->capacity; i++) {
    if (table->rows[i] != NULL) {
      free(table->rows[i]);
    }
  }
  free(table->rows);
}
