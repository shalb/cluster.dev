#!/usr/bin/env python3
"""Logging stuff."""
import logging


class CustomFormatter(logging.Formatter):
    """Logging Formatter for customize message for each error level."""

    formats = {
        logging.ERROR: '\nERROR: {msg}',
        logging.WARNING: 'WARNING: {msg}',
        logging.DEBUG: 'DEBUG: {module}: {lineno}: {msg}',
        'DEFAULT': '{msg}',
    }

    def format(self, record):  # noqa: WPS125
        """Change format based on log level."""
        log_fmt = self.formats.get(record.levelno, self.formats['DEFAULT'])
        formatter = logging.Formatter(log_fmt, style='{')
        return formatter.format(record)


logger = logging.getLogger(__name__)
logger.setLevel(logging.DEBUG)
logger_ch = logging.StreamHandler()
logger_ch.setLevel(logging.INFO)
logger_ch.setFormatter(CustomFormatter())
logger.addHandler(logger_ch)
