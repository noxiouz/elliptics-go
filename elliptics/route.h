/*
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 */

#ifndef __ELLIPTICS_ROUTE_H
#define __ELLIPTICS_ROUTE_H

#include "session.h"

#ifdef __cplusplus
extern "C" {
#endif

void session_get_routes(ell_session *session, void *result_array_context);

#ifdef __cplusplus
}
#endif

#endif // __ELLIPTICS_ROUTE_H
