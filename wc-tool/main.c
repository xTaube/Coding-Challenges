#include <ctype.h>
#include <locale.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <wchar.h>

void print_allowed_options() {
  printf("Only allowed:\n \t-c - return bytes "
         "count\n\t-m - returns characters count\n\t-w - returns words "
         "count\n\t-l - returns lines count\n");
}

enum Option { OPTION_C = 0, OPTION_W = 1, OPTION_L = 2, OPTION_M = 3 };

enum Option map_to_option(char *option) {
  if (strncmp(option, "-c", 2) == 0)
    return OPTION_C;
  if (strncmp(option, "-w", 2) == 0)
    return OPTION_W;
  if (strncmp(option, "-l", 2) == 0)
    return OPTION_L;
  if (strncmp(option, "-m", 2) == 0)
    return OPTION_M;

  printf("Unknown option %s.\n", option);
  print_allowed_options();
  exit(EXIT_FAILURE);
}

struct Stats {
  char *filename;
  uint64_t bytes_count;
  uint64_t lines_count;
  uint64_t words_count;
  uint64_t chars_count;
};

int main(int argc, char *argv[]) {
  int fd;
  struct Stats stats = {NULL, 0, 0, 0, 0};
  enum Option options[4];

  char buf[MB_CUR_MAX];
  size_t read_bytes;
  mbstate_t state;
  wchar_t wc;
  wchar_t prev_wc = 0;

  setlocale(LC_ALL, "");

  if (argc > 1 && strncmp(argv[argc - 1], "-", 1) != 0) {
    stats.filename = argv[argc - 1];
    FILE *fptr = fopen(stats.filename, "rb");

    if (fptr == NULL) {
      printf("File %s doesn't exist\n", stats.filename);
      exit(EXIT_FAILURE);
    }
    fd = fileno(fptr);
    argc--;
  } else
    fd = 0; // stdin

  uint8_t options_num = argc - 1;
  if (options_num > 4) {
    printf("Too many options passed.\n");
    print_allowed_options();
    exit(EXIT_FAILURE);
  }

  for (int i = 0; i < options_num; i++)
    options[i] = map_to_option(argv[i + 1]);

  while ((read_bytes = read(fd, &buf, 1)) > 0) {
    size_t l = mbrtowc(&wc, buf, read_bytes, &state);
    stats.bytes_count++;

    if (l == (size_t)-2) {
      continue;
    }
    stats.chars_count++;

    if (wc == L'\0') {
      memset(&state, 0, sizeof(mbstate_t));
    }

    if (wc == '\n') {
      stats.lines_count++;
    }
    if (isspace(wc) && !isspace(prev_wc))
      stats.words_count++;

    prev_wc = wc;
  }

  printf("\t");
  for (uint8_t i = 0; i < options_num; i++) {
    switch (options[i]) {
    case OPTION_C:
      printf("%lld\t", stats.bytes_count);
      break;
    case OPTION_L:
      printf("%lld\t", stats.lines_count);
      break;
    case OPTION_M:
      printf("%lld\t", stats.chars_count);
      break;
    case OPTION_W:
      printf("%lld\t", stats.words_count);
      break;
    }
  }

  if (options_num == 0)
    printf("%lld\t%lld\t%lld ", stats.lines_count, stats.words_count,
           stats.bytes_count);

  if (stats.filename != NULL) {
    printf("%s", stats.filename);
    close(fd);
  }

  printf("\n");
  return 0;
}
