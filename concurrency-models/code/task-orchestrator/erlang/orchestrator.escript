#!/usr/bin/env escript
%%! -noshell

main(_) ->
    SuccessTasks = [{profile, 30, ok}, {billing, 45, ok}, {risk, 20, ok}],
    TimeoutTasks = [{profile, 30, ok}, {billing, 45, ok}, {risk, 120, ok}],
    FailureTasks = [{profile, 30, ok}, {billing, 45, fail}, {risk, 80, ok}],
    print_summary(orchestrate("erlang-success", SuccessTasks, 100)),
    print_summary(orchestrate("erlang-timeout", TimeoutTasks, 70)),
    print_summary(orchestrate("erlang-worker-error", FailureTasks, 100)).

orchestrate(Scenario, Tasks, TimeoutMs) ->
    process_flag(trap_exit, true),
    Parent = self(),
    Pairs = [{spawn_link(fun() -> run_task(Parent, Name, Delay, Mode) end), Name} || {Name, Delay, Mode} <- Tasks],
    Deadline = erlang:monotonic_time(millisecond) + TimeoutMs,
    collect(Scenario, Pairs, [], [], [], undefined, Deadline).

run_task(Parent, Name, Delay, Mode) ->
    timer:sleep(Delay),
    case Mode of
        ok -> Parent ! {result, self(), Name, ok};
        fail -> exit({task_error, Name})
    end.

collect(Scenario, [], Completed, Failed, Canceled, Error, _Deadline) ->
    #{scenario => Scenario,
      completed => lists:sort(Completed),
      failed => lists:sort(Failed),
      canceled => lists:sort(Canceled),
      error => Error};
collect(Scenario, Pairs, Completed, Failed, Canceled, Error, Deadline) ->
    Wait = max(0, Deadline - erlang:monotonic_time(millisecond)),
    receive
        {result, Pid, Name, ok} ->
            Rest = lists:keydelete(Pid, 1, Pairs),
            collect(Scenario, Rest, [Name | Completed], Failed, Canceled, Error, Deadline);
        {'EXIT', Pid, {task_error, Name}} ->
            Rest = lists:keydelete(Pid, 1, Pairs),
            cancel(Rest),
            CanceledNames = [N || {_P, N} <- Rest],
            NewError = case Error of undefined -> {task_error, Name}; _ -> Error end,
            collect(Scenario, [], Completed, [Name | Failed], CanceledNames ++ Canceled, NewError, Deadline);
        {'EXIT', Pid, killed} ->
            case proplists:get_value(Pid, Pairs, undefined) of
                undefined ->
                    collect(Scenario, Pairs, Completed, Failed, Canceled, Error, Deadline);
                Name ->
                    Rest = lists:keydelete(Pid, 1, Pairs),
                    collect(Scenario, Rest, Completed, Failed, [Name | Canceled], Error, Deadline)
            end;
        {'EXIT', Pid, normal} ->
            Rest = lists:keydelete(Pid, 1, Pairs),
            collect(Scenario, Rest, Completed, Failed, Canceled, Error, Deadline)
    after Wait ->
        cancel(Pairs),
        CanceledNames = [N || {_P, N} <- Pairs],
        TimeoutError = case Error of undefined -> timeout; _ -> Error end,
        collect(Scenario, [], Completed, Failed, CanceledNames ++ Canceled, TimeoutError, Deadline)
    end.

cancel(Pairs) ->
    [exit(Pid, kill) || {Pid, _Name} <- Pairs],
    ok.

print_summary(Summary) ->
    io:format("scenario=~s~n", [maps:get(scenario, Summary)]),
    io:format("state_owner=each process owns its local state; parent aggregates only messages~n"),
    io:format("waiting_owner=mailbox receive and after timeout define waiting boundary in the parent process~n"),
    io:format("failure_boundary=linked worker exit is received by parent; parent kills sibling processes explicitly~n"),
    io:format("completed=~p~n", [maps:get(completed, Summary)]),
    io:format("failed=~p~n", [maps:get(failed, Summary)]),
    io:format("canceled=~p~n", [maps:get(canceled, Summary)]),
    Error = maps:get(error, Summary),
    case Error of
        undefined -> io:format("error=<nil>~n");
        _ -> io:format("error=~p~n", [Error])
    end,
    io:format("---~n").
