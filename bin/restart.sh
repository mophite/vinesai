#!/usr/bin/env bash
dir=`pwd`

run() {
  echo "$dir";\
	for d in $(ls ./$1); do
		echo "run $1/$d"
		pushd $dir/$1/$d >/dev/null
	    kill -9 $(ps -ef|grep "\./$d" |awk '$0 !~/grep/ {print $2}' |tr -s '\n' ' ')
		nohup ./$d &
		popd >/dev/null
	done

  kill -9 $(ps -ef|grep "\./srv" |awk '$0 !~/grep/ {print $2}' |tr -s '\n' ' ')

	ps -ef|grep api.
	ps -ef|grep srv.
}

run cmd;\
exit