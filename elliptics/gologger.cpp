#include "gologger.h"
#include "session.h"

using namespace ioremap;

extern "C" {

#include "_cgo_export.h"

typedef void logger_function_t(void *, int, char *);

class logger_interface : public elliptics::logger_interface {
	public:
		logger_interface(void *func, void *priv) : m_func(func), m_priv(priv) {}
		~logger_interface() {}

		void log(const int level, const char *msg) {
			::GoLog(m_func, m_priv, level, const_cast<char *>(msg));
		}

	private:
		void *m_func, *m_priv;
};

ell_node *gologger_create(void *func, void *priv, const int level)
{
	try {
		logger_interface *intf = new logger_interface(func, priv);

		try {
			ioremap::elliptics::logger gol(intf, level);
			return new ioremap::elliptics::node(gol);
		} catch (...) {
			delete intf;
		}
	} catch (...) {
	}

	return NULL;
}

} // extern "C"
