/*
 * 2014+ Copyright (c) Evgeniy Polyakov <zbr@ioremap.net>
 * All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 */

#ifndef __ELLIPTICS_STAT_H
#define __ELLIPTICS_STAT_H


#include "session.h"

#include <elliptics/packet.h>

struct go_stat_result {
	const struct dnet_cmd		*cmd;
	const struct dnet_addr		*addr;
	const char			*stat_data;
	const size_t			stat_size;
};

#ifdef __cplusplus
extern "C" {
#endif

void session_get_stats(ell_session *session, context_t on_chunk_context, context_t final_context, uint64_t categories);

#ifdef __cplusplus
}
#endif

#endif // __ELLIPTICS_STAT_H
