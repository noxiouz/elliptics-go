/*
* 2013+ Copyright (c) Anton Tyurin <noxiouz@yandex.ru>
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

#include "logger.h"
#include <ios>
#include <errno.h>
#include <iostream>
#include <memory>

#include <elliptics/session.hpp>

using namespace ioremap;

extern "C" {

ell_file_logger* 
new_file_logger(const char *file) {
	try {
		elliptics::file_logger* l = new elliptics::file_logger(file);
		return (ell_file_logger*)l;
	} catch (const std::ios_base::failure& e) {
		errno = ENOENT;
		std::cerr << e.what() << std::endl;
		return NULL;
	}
}

void                 
file_logger_log(ell_file_logger *fl, int level, const char *msg) {
    ioremap::elliptics::file_logger *x = (ioremap::elliptics::file_logger*)fl;
	x->log(level, msg);
}

int
file_logger_get_level(ell_file_logger *fl) {
	ioremap::elliptics::file_logger *x = (ioremap::elliptics::file_logger*)fl;
	return x->get_log_level();
}

void 
delete_file_logger(ell_file_logger *fl) {
	delete (ioremap::elliptics::file_logger*)fl;
}

} // extern "C"
