"use client";
import React, { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

type User = { username: string };
type SearchResult = { message: string; receiver: User };
type WSMessage = { from: string; type: string };

const Home = () => {
  const router = useRouter();
  const API_BASE = "http://localhost:8080";
  const socketRef = useRef<WebSocket | null>(null);

  const [username, setUsername] = useState("");
  const [usernameInput, setUsernameInput] = useState("");
  const [searchedUser, setSearchedUser] = useState<SearchResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [friendRequests, setFriendRequests] = useState<User[]>([]);
  const [friendList, setFriendList] = useState<User[]>([]);

  /* ---------------- RESTORE SESSION ---------------- */
  useEffect(() => {
    const storedUsername = sessionStorage.getItem("username") || "";
    setUsername(storedUsername);
  }, []);

  /* ---------------- WEBSOCKET CONNECT ---------------- */
  useEffect(() => {
    const token = sessionStorage.getItem("access_token");
    if (!token) return;

    const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
    
    ws.onopen = () => console.log("WebSocket connected");

    ws.onmessage = (event) => {
      const msg: WSMessage = JSON.parse(event.data);

      if (msg.type === "friend_request") {
        setFriendRequests((prev) => {
          if (!prev.find((u) => u.username === msg.from)) {
            return [...prev, { username: msg.from }];
          }
          return prev;
        });
      }
    };

    ws.onclose = () => console.log("WebSocket disconnected");

    socketRef.current = ws;
    return () => ws.close();
  }, []);

  /* ---------------- AUTH FETCH ---------------- */
  async function protectedFetch(url: string) {
    const accessToken = sessionStorage.getItem("access_token");
    const refreshToken = sessionStorage.getItem("refresh_token");

    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    if (res.ok) return res.json();

    const refreshRes = await fetch(`${API_BASE}/refresh`, {
      method: "POST",
      headers: { Authorization: `Bearer ${refreshToken}` },
    });

    if (!refreshRes.ok) {
      sessionStorage.clear();
      router.push("/");
      throw new Error("Session expired");
    }

    const data = await refreshRes.json();
    sessionStorage.setItem("access_token", data.access_token);

    socketRef.current?.close();
    socketRef.current = new WebSocket(
      `ws://localhost:8080/ws?token=${data.access_token}`
    );

    const retry = await fetch(url, {
      headers: { Authorization: `Bearer ${data.access_token}` },
    });

    return retry.json();
  }

  /* ---------------- SEARCH USER ---------------- */
  async function searchUser() {
    setError(null);
    setSearchedUser(null);

    if (!usernameInput || usernameInput === username) {
      setError("Invalid username");
      return;
    }

    try {
      const res: SearchResult = await protectedFetch(
        `${API_BASE}/user/${usernameInput}`
      );
      setSearchedUser(res);
      setUsernameInput("");
      console.log(res);
    } catch (err: any) {
      setError(err.message);
    }
  }

  /* ---------------- SEND FRIEND REQUEST ---------------- */
  function sendFriendRequest(toUser: User) {
    socketRef.current?.send(
      JSON.stringify({ type: "friend_request", to: toUser.username, content: "" })
    );
    setSearchedUser((prev) =>
      prev ? { ...prev, message: "pending" } : null
    );
  }

  /* ---------------- ACCEPT / REJECT ---------------- */
  function handleRequest(fromUser: User, accept: boolean) {
    setFriendRequests((prev) =>
      prev.filter((u) => u.username !== fromUser.username)
    );
    if (accept) setFriendList((prev) => [...prev, fromUser]);
  }

  return (
    <div className="min-h-screen p-4 flex">
      {/* Left side panel */}
      <div className="flex flex-col gap-4 w-80">
        {/* Top-left search */}
        <div className="flex flex-col gap-4">
          <h1 className="text-xl font-bold">Welcome {username}</h1>

          <div className="flex gap-2">
            <input
              value={usernameInput}
              onChange={(e) => setUsernameInput(e.target.value)}
              placeholder="Search username"
              className="border rounded px-2 py-1 flex-1"
            />
            <button
              onClick={searchUser}
              className="bg-blue-500 text-white px-4 rounded"
            >
              Search
            </button>
          </div>

          {error && <p className="text-sm text-red-500">{error}</p>}

          {searchedUser && (
            <div className="flex justify-between items-center border p-2 rounded">
              <span>{searchedUser.receiver.username}</span>
              <div className="flex gap-2">
                {searchedUser.message === "available" && (
                  <button
                    onClick={() => sendFriendRequest(searchedUser.receiver)}
                    className="bg-green-500 text-white px-3 rounded"
                  >
                    Send Friend Request
                  </button>
                )}

                {searchedUser.message === "pending" && (
                  <button
                    disabled
                    className="bg-gray-400 text-white px-3 rounded cursor-not-allowed"
                  >
                    Pending Friend Request
                  </button>
                )}

                {/* accepted -> no button */}
                {searchedUser.message === "accepted" && null}

                <button
                  onClick={() => setSearchedUser(null)}
                  className="bg-red-500 text-white px-2 rounded"
                >
                  X
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Friend Request Inbox */}
        <div className="border rounded p-2 h-64 overflow-y-auto flex flex-col gap-2">
          <h2 className="font-semibold mb-2">Friend Request Inbox</h2>
          {friendRequests.map((req) => (
            <div
              key={req.username}
              className="flex justify-between items-center border p-2 rounded"
            >
              <span>{req.username}</span>
              <div className="flex gap-2">
                <button
                  onClick={() => handleRequest(req, true)}
                  className="bg-green-500 text-white px-2 rounded"
                >
                  Accept
                </button>
                <button
                  onClick={() => handleRequest(req, false)}
                  className="bg-red-500 text-white px-2 rounded"
                >
                  Reject
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Friend List */}
        <div className="border rounded p-2 h-64 overflow-y-auto flex flex-col gap-2">
          <h2 className="font-semibold mb-2">Friend List</h2>
          {friendList.map((f) => (
            <div key={f.username} className="border p-2 rounded">
              {f.username}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Home;
