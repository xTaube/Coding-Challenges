// Link to the challenge
// https://codingchallenges.fyi/challenges/challenge-huffman/

#include <assert.h>
#include <complex.h>
#include <locale.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <wchar.h>

#define START_UTF8_CHAR_TABLE_SIZE 1048

int main(int argc, char *argv[]) {
  char buf[MB_CUR_MAX];
  mbstate_t state;
  size_t read_bytes;

  if (argc < 2) {
    printf("File is missing.");
  }

  char *filename = argv[argc - 1];
  FILE *fptr = fopen(filename, "rb");

  if (fptr == NULL) {
    printf("File %s doesn't exists.\n", filename);
    exit(EXIT_FAILURE);
  }

  int fd = fileno(fptr);
  setlocale(LC_ALL, "");

  struct UTF8CharTable table = {NULL, START_UTF8_CHAR_TABLE_SIZE, 0};
  table.rows = (struct UTF8CharRow **)malloc(sizeof(struct UTF8CharRow *) *
                                             table.capacity);

  for (uint32_t i = 0; i < table.capacity; i++) {
    table.rows[i] = NULL;
  }

  wchar_t wc;
  while ((read_bytes = read(fd, buf, 1)) > 0) {
    size_t len = mbrtowc(&wc, buf, read_bytes, &state);

    if (len == (size_t)-2)
      continue;
    if (wc == L'\0')
      memset(&state, 0, sizeof(mbstate_t));

    increment_counter(&table, wc);
  }

  // check known occureces from test.txt
  assert(get_count(&table, 'X') == 333);
  assert(get_count(&table, 't') == 223000);

  free_utf8_char_table(&table);
  return 0;
}
