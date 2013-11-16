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

#include <elliptics/session.hpp>

#include "logger.h"

#include <errno.h>
#include <ios>
#include <iostream>
#include <memory>

using namespace ioremap;

extern "C" {

ell_file_logger* 
new_file_logger(const char *file) {
	try {
		elliptics::file_logger* l = new elliptics::file_logger(file);
		return (ell_file_logger*)l;
	} catch (const std::ios_base::failure& e) {
		return NULL;
	}
}

void                 
file_logger_log(ell_file_logger *fl, int level, const char *msg) {
    elliptics::file_logger *efl = (elliptics::file_logger*)fl;
	efl->log(level, msg);
}

int
file_logger_get_level(ell_file_logger *fl) {
	elliptics::file_logger *efl = (elliptics::file_logger*)fl;
	return efl->get_log_level();
}

void 
delete_file_logger(ell_file_logger *fl) {
	delete (elliptics::file_logger*)fl;
}

} // extern "C"
