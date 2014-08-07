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

#include <errno.h>
#include <ios>
#include <iostream>
#include <memory>

using namespace ioremap;

extern "C" {

ell_file_logger *new_file_logger(const char *file, int level)
{
	try {
		elliptics::file_logger *l = new elliptics::file_logger(file, level);
		return l;
	} catch(const std::exception & e) {
		std::cerr << "go: file-logger-exception: " << e.what() << std::endl;
		return NULL;
	}
}

void file_logger_log(ell_file_logger *fl, int level, const char *msg)
{
    //ToDo: change log level
    auto record = fl->open_record(::dnet_log_level::DNET_LOG_DATA);
}

int file_logger_get_level(ell_file_logger *fl)
{
	return fl->verbosity();
}

void delete_file_logger(ell_file_logger *fl)
{
	delete fl;
}

} // extern "C"
