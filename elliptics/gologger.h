#ifndef __GOLOGGER_H
#define __GOLOGGER_H

#include "node.h"

#ifdef __cplusplus
extern "C" {
#endif

ell_node *gologger_create(void *func, void *priv, const int level);

#ifdef __cplusplus
}
#endif

#endif /* __GOLOGGER_H */
