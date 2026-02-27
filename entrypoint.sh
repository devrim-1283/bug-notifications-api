#!/bin/sh
set -e

case "${MODE}" in
  api)
    echo "Starting API server on port ${PORT}..."
    exec api
    ;;
  worker)
    echo "Starting Worker..."
    exec worker
    ;;
  all|*)
    echo "Starting API server on port ${PORT} and Worker..."
    worker &
    WORKER_PID=$!

    trap 'kill $WORKER_PID 2>/dev/null; wait $WORKER_PID 2>/dev/null' TERM INT

    api &
    API_PID=$!

    wait $API_PID $WORKER_PID
    ;;
esac
