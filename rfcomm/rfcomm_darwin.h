int q30_open(const char *addr, void **handle);
int q30_write(void *handle, const void *p, int n);
int q30_read(void *handle, void *out, int cap, double timeout);
void q30_close(void *handle);
int q30_connect(const char *addr);
int q30_disconnect(const char *addr);
int q30_connected(const char *addr);
char *q30_find(const char *name);
