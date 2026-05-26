/*
 * Registers the jq extension with every SQLite connection opened by the
 * sqlite3 shell, without requiring .load.
 *
 * Called via the SQLITE_SHELL_INIT_PROC hook, which the shell invokes after
 * all sqlite3_config() calls but before opening any database. This avoids the
 * "attempt to configure SQLite after initialization" warning that would occur
 * if we used __attribute__((constructor)) and called sqlite3_auto_extension
 * before main() ran.
 *
 * sqlite3_auto_extension is deprecated on macOS 10.10+ for sandboxed apps;
 * the warning is suppressed here because this binary is not sandboxed.
 */
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"

#include "../sqlite/sqlite3.h"

extern int sqlite3_extension_init(sqlite3 *, char **, const sqlite3_api_routines *);

void register_jq_extension(void) {
    sqlite3_auto_extension((void (*)(void))sqlite3_extension_init);
}

#pragma clang diagnostic pop
