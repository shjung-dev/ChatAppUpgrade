"use client";
import React, { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";

type User = { username: string };
type SearchResult = { message: string; receiver: User };
type Message = {
  SenderUserName: string;
  Content: string;
  CreatedAt: string;
};
type Conversation = {
  ID: string;
  ConversationID: string;
  ConversationName: string | null;
  Participants: string[];
  CreatedAt: string;
  LastMessageAt: string | null;
};
type WSMessage = {
  from: string;
  type: string;
  friends?: { friendusername: string }[];
  convo?: Conversation;
  message?: Message;
  convoAndMessages?: { conversation: Conversation; Messages: Message[] }[];
  convoID?: string;
  groupName?: string;
  members?: string[];
  messageContent?: string;
};

const Home = () => {
  const router = useRouter();
  const API_BASE = "https://upgradedchatappservice.onrender.com";
  const socketRef = useRef<WebSocket | null>(null);
  const chatEndRef = useRef<HTMLDivElement | null>(null);

  const [username, setUsername] = useState("");
  const [usernameInput, setUsernameInput] = useState("");
  const [searchedUser, setSearchedUser] = useState<SearchResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [friendRequests, setFriendRequests] = useState<User[]>([]);
  const [friendList, setFriendList] = useState<User[]>([]);

  /* ---------------- CHAT UI STATE ---------------- */
  const [chatUsers, setChatUsers] = useState<Conversation[]>([]);
  const [activeChat, setActiveChat] = useState<Conversation | null>(null);
  const [chatInput, setChatInput] = useState("");
  const [chatMessages, setChatMessages] = useState<{
    [convoID: string]: Message[];
  }>({});
  const [messages, setMessages] = useState<Message[]>([]);

  /* ---------------- TEMPORARY CHAT STATE ---------------- */
  const [temporaryChats, setTemporaryChats] = useState<Conversation[]>([]);

  /* ---------------- GROUP CHAT MODAL ---------------- */
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [groupName, setGroupName] = useState("");
  const [selectedFriends, setSelectedFriends] = useState<Set<string>>(
    new Set()
  );

  /* ---------------- RESTORE SESSION ---------------- */
  useEffect(() => {
    const storedUsername = sessionStorage.getItem("username") || "";
    setUsername(storedUsername);
  }, []);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  /* ---------------- SEND MESSAGE ---------------- */
  async function sendMessage() {
    if (!chatInput || !activeChat) return;

    const messageObj: Message = {
      SenderUserName: username,
      Content: chatInput,
      CreatedAt: new Date().toISOString(),
    };

    const messagePayload: WSMessage = {
      from: username,
      type: "message",
      convoID: activeChat.ConversationID.startsWith("temp-")
        ? "" // backend will create conversation ID
        : activeChat.ConversationID,
      groupName: activeChat.ConversationName || "",
      members: activeChat.Participants.filter((u) => u !== username),
      messageContent: chatInput,
    };

    // Optimistically add message to chatMessages
    setChatMessages((prev) => ({
      ...prev,
      [activeChat.ConversationID]: [
        ...(prev[activeChat.ConversationID]?.filter(
          (m) =>
            !(
              m.SenderUserName === messageObj.SenderUserName &&
              m.Content === messageObj.Content
            )
        ) || []),
        messageObj,
      ],
    }));

    // If this is a first message to a "temp" chat, add it to chatUsers immediately
    if (activeChat.ConversationID.startsWith("temp-")) {
      const newChat: Conversation = {
        ...activeChat,
        ConversationID: activeChat.ConversationID, // temp ID until backend replaces
        LastMessageAt: messageObj.CreatedAt,
      };

      setChatUsers((prev) => [newChat, ...prev]); // add to top of chat list
    }

    // Update active chat messages
    setMessages((prev) => [...(prev || []), messageObj]);
    setChatInput("");

    // Send to backend
    socketRef.current?.send(JSON.stringify(messagePayload));
  }

  /* ---------------- HANDLE INCOMING WEBSOCKET MESSAGE ---------------- */
  useEffect(() => {
    const token = sessionStorage.getItem("access_token");
    if (!token) return;

    const ws = new WebSocket(
      `wss://upgradedchatappservice.onrender.com/ws?token=${token}`
    );
    socketRef.current = ws;

    ws.onopen = () => console.log("WebSocket connected");

    ws.onmessage = (event) => {
      const msg: WSMessage = JSON.parse(event.data);

      if (msg.type === "friend_request") {
        setFriendRequests((prev) =>
          prev.find((u) => u.username === msg.from)
            ? prev
            : [...prev, { username: msg.from }]
        );
      }

      if (msg.type === "friend_list_update" && msg.friends) {
        const updatedFriends = msg.friends.map((f) => ({
          username: f.friendusername,
        }));
        setFriendList(updatedFriends as User[]);
      }

      if (msg.type === "message" && msg.convo && msg.message) {
        const convoObj: Conversation = msg.convo; // backend conversation
        const messageObj: Message = msg.message;

        // Remove any temp chat with same participants (if exists)
        let replacedTempChat = false;
        setChatUsers((prev) => {
          // Check if a temp chat exists with the same participants
          const tempIndex = prev.findIndex(
            (c) =>
              c.ConversationID.startsWith("temp-") &&
              c.Participants.length === convoObj.Participants.length &&
              c.Participants.every((p) => convoObj.Participants.includes(p))
          );

          let newUsers = [...prev];

          if (tempIndex !== -1) {
            // Replace the temp chat with backend chat
            newUsers[tempIndex] = {
              ...convoObj,
              LastMessageAt: messageObj.CreatedAt,
            };
            replacedTempChat = true;
          } else {
            // Add backend chat if not exists
            if (
              !prev.find((c) => c.ConversationID === convoObj.ConversationID)
            ) {
              newUsers = [
                ...newUsers,
                { ...convoObj, LastMessageAt: messageObj.CreatedAt },
              ];
            }
          }

          return newUsers;
        });

        // Add message to chatMessages
        setChatMessages((prev) => ({
          ...prev,
          [convoObj.ConversationID]: [
            ...(prev[convoObj.ConversationID]?.filter(
              (m) =>
                !(
                  m.SenderUserName === messageObj.SenderUserName &&
                  m.Content === messageObj.Content
                )
            ) || []),
            messageObj,
          ],
        }));

        // Update active chat: replace temp chat with real backend chat if needed
        setActiveChat((prev) => {
          if (!prev) return prev;

          if (prev.ConversationID === convoObj.ConversationID) {
            return convoObj; // already backend chat, just update
          }

          // If previous active chat was temp with same participants, replace with backend chat
          if (
            prev.ConversationID.startsWith("temp-") &&
            prev.Participants.length === convoObj.Participants.length &&
            prev.Participants.every((p) => convoObj.Participants.includes(p))
          ) {
            return convoObj;
          }

          return prev; // otherwise keep current active chat
        });

        // Update messages in activeChat if active
        setMessages((prevMsgs) => {
          if (
            activeChat &&
            ((activeChat.ConversationID.startsWith("temp-") &&
              activeChat.Participants.length === convoObj.Participants.length &&
              activeChat.Participants.every((p) =>
                convoObj.Participants.includes(p)
              )) ||
              activeChat.ConversationID === convoObj.ConversationID)
          ) {
            return [
              ...(prevMsgs.filter(
                (m) =>
                  !(
                    m.SenderUserName === messageObj.SenderUserName &&
                    m.Content === messageObj.Content
                  )
              ) || []),
              messageObj,
            ];
          }
          return prevMsgs;
        });
      }

      if (msg.type === "allMessages" && msg.convoAndMessages) {
        const newChats = msg.convoAndMessages.map((item) => item.conversation);

        const newChatMessages: { [convoID: string]: Message[] } = {};
        msg.convoAndMessages.forEach((item) => {
          const sortedMsgs = item.Messages.sort(
            (a, b) =>
              new Date(a.CreatedAt).getTime() - new Date(b.CreatedAt).getTime()
          );
          newChatMessages[item.conversation.ConversationID] = sortedMsgs;
        });

        newChats.sort((a, b) => {
          const aTime = a.LastMessageAt
            ? new Date(a.LastMessageAt).getTime()
            : 0;
          const bTime = b.LastMessageAt
            ? new Date(b.LastMessageAt).getTime()
            : 0;
          return bTime - aTime;
        });

        setChatUsers(newChats);
        setChatMessages(newChatMessages);
      }
    };
  }, [activeChat]);

  /* ---------------- OPEN CHAT ---------------- */
  function openChat(convo: Conversation) {
    setActiveChat(convo);
    setMessages(chatMessages[convo.ConversationID] || []);
  }

  /* ---------------- GENERIC AUTH FETCH ---------------- */
  async function protectedFetch(url: string, options: RequestInit = {}) {
    const accessToken = sessionStorage.getItem("access_token");
    const refreshToken = sessionStorage.getItem("refresh_token");

    const res = await fetch(url, {
      ...options,
      headers: {
        ...(options.headers || {}),
        Authorization: `Bearer ${accessToken}`,
      },
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
    sessionStorage.setItem("refresh_token", data.refresh_token);

    socketRef.current?.close();
    socketRef.current = new WebSocket(
      `ws://localhost:8080/ws?token=${data.access_token}`
    );

    const retry = await fetch(url, {
      ...options,
      headers: {
        ...(options.headers || {}),
        Authorization: `Bearer ${data.access_token}`,
      },
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
    } catch (err: any) {
      setError(err.message);
    }
  }

  /* ---------------- SEND FRIEND REQUEST ---------------- */
  function sendFriendRequest(toUser: User) {
    socketRef.current?.send(
      JSON.stringify({ type: "friend_request", to: toUser.username })
    );
    setSearchedUser((prev) => (prev ? { ...prev, message: "pending" } : null));
  }

  async function handleRequest(fromUser: User, accept: boolean) {
    if (accept) {
      setFriendRequests((prev) =>
        prev.filter((u) => u.username !== fromUser.username)
      );
      setFriendList((prev) =>
        prev.find((f) => f.username === fromUser.username)
          ? prev
          : [...prev, fromUser]
      );
    } else {
      setFriendRequests((prev) =>
        prev.filter((u) => u.username !== fromUser.username)
      );
    }

    try {
      if (accept) {
        await protectedFetch(`${API_BASE}/accept/${fromUser.username}`, {
          method: "POST",
        });
        socketRef.current?.send(JSON.stringify({ type: "friend_list_update" }));
      } else {
        await protectedFetch(`${API_BASE}/reject/${fromUser.username}`, {
          method: "POST",
        });
      }
    } catch (err) {
      console.error("Failed to handle request:", err);
    }
  }

  async function removeFriend(friend: User) {
    try {
      setFriendList((prev) =>
        prev.filter((f) => f.username !== friend.username)
      );
      await protectedFetch(`${API_BASE}/remove/${friend.username}`, {
        method: "POST",
      });
      socketRef.current?.send(JSON.stringify({ type: "friend_list_update" }));
    } catch (err) {
      console.error("Failed to remove friend:", err);
      setFriendList((prev) => [...prev, friend]);
    }
  }

  /* ---------------- CREATE GROUP CHAT ---------------- */
  function createGroupChat() {
    if (!groupName || selectedFriends.size === 0) return;

    const members = Array.from(selectedFriends);
    socketRef.current?.send(
      JSON.stringify({
        type: "message",
        from: username,
        groupName: groupName,
        members: members,
        messageContent: `${username} created the group chat ${groupName}`,
      })
    );

    setShowGroupModal(false);
    setGroupName("");
    setSelectedFriends(new Set());
  }

  const getChatName = (convo: Conversation) => {
    if (convo.ConversationName) return convo.ConversationName;
    return convo.Participants.find((p) => p !== username) || "Unknown";
  };

  /* ---------------- HANDLE FRIEND CLICK ---------------- */
  const handleFriendClick = (friend: User) => {
    // Check if a backend 1:1 chat exists
    const existingChat = chatUsers.find(
      (c) =>
        c.Participants.length === 2 && // only 1:1 chats
        c.Participants.includes(friend.username)
    );
    if (existingChat) {
      openChat(existingChat);
      return;
    }

    // If backend chat doesn't exist, just open temporary UI (not in chat list)
    const tempChat: Conversation = {
      ID: "",
      ConversationID: `temp-${friend.username}-${Date.now()}`,
      ConversationName: null,
      Participants: [username, friend.username],
      CreatedAt: new Date().toISOString(),
      LastMessageAt: null,
    };

    // Open temporary chat UI
    setActiveChat(tempChat);
    setMessages([]);
  };

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-linear-to-br from-slate-50 to-slate-100 dark:from-zinc-900 dark:to-zinc-950">
      {/* ================= SIDEBAR ================= */}
      <aside className="hidden md:flex w-80 flex-col border-r bg-white/80 dark:bg-zinc-900/80 backdrop-blur">
        <div className="p-4 border-b">
          <h2 className="text-lg font-semibold">Welcome, {username}</h2>

          <input
            type="text"
            placeholder="Search username..."
            value={usernameInput}
            onChange={(e) => setUsernameInput(e.target.value)}
            className="mt-3 w-full rounded-xl border px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500"
          />

          <button
            onClick={searchUser}
            className="mt-2 w-full rounded-xl bg-blue-600 py-2 text-sm font-medium text-white transition hover:bg-blue-700 active:scale-95"
          >
            Search
          </button>

          {error && <p className="mt-2 text-sm text-red-500">{error}</p>}

          {searchedUser && (
            <div className="mt-3 flex items-center justify-between rounded-lg border p-2">
              <span className="text-sm font-medium">
                {searchedUser.receiver.username}
              </span>
              {searchedUser.message === "pending" ? (
                <span className="text-xs text-gray-400">Pending</span>
              ) : (
                <button
                  onClick={() => sendFriendRequest(searchedUser.receiver)}
                  className="rounded-lg bg-green-500 px-3 py-1 text-xs text-white hover:bg-green-600"
                >
                  Add
                </button>
              )}
            </div>
          )}
        </div>

        {/* Friend Requests */}
        <div className="p-4 border-b">
          <h3 className="mb-2 text-sm font-semibold text-gray-500">
            Friend Requests
          </h3>
          {friendRequests.length === 0 && (
            <p className="text-sm text-gray-400">No requests</p>
          )}
          {friendRequests.map((req) => (
            <div
              key={req.username}
              className="mb-2 flex items-center justify-between rounded-lg p-2 hover:bg-gray-100 dark:hover:bg-zinc-800"
            >
              <span className="text-sm">{req.username}</span>
              <div className="flex gap-1">
                <button
                  onClick={() => handleRequest(req, true)}
                  className="rounded bg-green-500 px-2 py-1 text-xs text-white"
                >
                  Accept
                </button>
                <button
                  onClick={() => handleRequest(req, false)}
                  className="rounded bg-red-500 px-2 py-1 text-xs text-white"
                >
                  Reject
                </button>
              </div>
            </div>
          ))}
        </div>

        {/* Friends */}
        <div className="flex-1 overflow-y-auto p-4">
          <h3 className="mb-2 text-sm font-semibold text-gray-500">Friends</h3>
          {friendList.length === 0 && (
            <p className="text-sm text-gray-400">No friends</p>
          )}
          {friendList.map((friend) => (
            <div
              key={friend.username}
              onClick={() => handleFriendClick(friend)}
              className="mb-2 flex items-center justify-between rounded-xl p-2 cursor-pointer transition hover:bg-gray-100 dark:hover:bg-zinc-800"
            >
              <span className="text-sm font-medium">{friend.username}</span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  removeFriend(friend);
                }}
                className="text-xs text-red-500 hover:underline"
              >
                Remove
              </button>
            </div>
          ))}
        </div>

        <div className="p-4 border-t">
          <button
            onClick={() => setShowGroupModal(true)}
            className="w-full rounded-xl bg-purple-600 py-2 text-sm font-medium text-white transition hover:bg-purple-700 active:scale-95"
          >
            Create Group Chat
          </button>
        </div>
      </aside>

      {/* ================= CHAT LIST ================= */}
      <section className="w-full md:w-80 border-r bg-white/70 dark:bg-zinc-900/70 backdrop-blur">
        <h3 className="border-b p-4 font-semibold">Chats</h3>
        <div className="h-full overflow-y-auto">
          {chatUsers.map((chat) => (
            <div
              key={chat.ConversationID}
              onClick={() => openChat(chat)}
              className={`cursor-pointer border-b p-3 transition hover:bg-gray-100 dark:hover:bg-zinc-800 ${
                activeChat?.ConversationID === chat.ConversationID
                  ? "bg-blue-50 dark:bg-zinc-800"
                  : ""
              }`}
            >
              <p className="font-medium">{getChatName(chat)}</p>
              <p className="truncate text-sm text-gray-500">
                {chatMessages[chat.ConversationID]?.slice(-1)[0]?.Content || ""}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* ================= ACTIVE CHAT ================= */}
      <main className="flex flex-1 flex-col">
        {activeChat ? (
          <>
            <div className="border-b p-4 font-semibold">
              {getChatName(activeChat)}
            </div>

            <div className="flex-1 space-y-3 overflow-y-auto p-4">
              {messages.length === 0 ? (
                <div className="flex h-full items-center justify-center text-gray-400">
                  No messages yet
                </div>
              ) : (
                messages.map((msg, idx) => {
                  const isMe = msg.SenderUserName === username;
                  return (
                    <div
                      key={idx}
                      className={`flex ${
                        isMe ? "justify-end" : "justify-start"
                      }`}
                    >
                      <div
                        className={`max-w-xs rounded-2xl px-4 py-2 text-sm shadow ${
                          isMe
                            ? "bg-blue-600 text-white"
                            : "bg-gray-200 dark:bg-zinc-800"
                        }`}
                      >
                        {msg.Content}
                      </div>
                    </div>
                  );
                })
              )}
              <div ref={chatEndRef} />
            </div>

            <div className="flex gap-2 border-t p-4">
              <input
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && sendMessage()}
                placeholder="Type a message..."
                className="flex-1 rounded-xl border px-4 py-2 focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={sendMessage}
                className="rounded-xl bg-blue-600 px-5 py-2 text-white transition hover:bg-blue-700 active:scale-95"
              >
                Send
              </button>
            </div>
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center text-gray-400">
            Select a chat to start messaging
          </div>
        )}
      </main>
    </div>
  );
};

export default Home;
