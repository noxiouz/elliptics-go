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

#ifndef LOGGER_H
#define LOGGER_H


#ifdef __cplusplus
#include <elliptics/session.hpp>
typedef ioremap::elliptics::file_logger ell_file_logger;
extern "C" {
#else
typedef void ell_file_logger;
#endif


ell_file_logger*
new_file_logger(const char *file);

void
delete_file_logger(ell_file_logger *fl);

void                 
file_logger_log(ell_file_logger *logger, int level, const char *msg);

int
file_logger_get_level(ell_file_logger *fl);

#ifdef __cplusplus 
}
#endif

#endif
