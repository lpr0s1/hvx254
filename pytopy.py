import socket
import threading
import random
import string
import os
import sys
import time
import signal
import queue

def kjl(signum, frame):
    os.system('clear')
    os.system('cls')
    os.system('clear')
    os.system('cls')
    sys.exit(0)
signal.signal(signal.SIGINT, kjl)
os.system("clear")
os.system('cls')
try:
    from art import tprint
except ImportError:
    def tprint(text):
        print(f"*** {text.upper()} ***")

PYTOPY_VERSION = "1.0"

class User:
    def __init__(self, conn, addr, pseudo, uid):
        self.conn = conn
        self.addr = addr
        self.pseudo = pseudo
        self.uid = uid

class P2PNode:
    def __init__(self):
        self.host = "0.0.0.0"
        self.port = 9999
        self.users = {}
        self.lock = threading.Lock()
        self.running = True
        self.local_uid = None
        self.local_pseudo = None
        self.sock = None 
        self.client_sock = None
        self.current_node_users = 0
        
        
        self.input_queue = queue.Queue()
        self.is_chatting = False 

    def generate_uid(self):
        return ''.join(random.choices(string.ascii_letters + string.digits, k=8))

    def choose_identity(self):
        tprint("pytopy")
        self.local_pseudo = input("Choisissez votre pseudo : ").strip()
        if not self.local_pseudo:
            self.local_pseudo = "Anonyme"
        self.local_uid = self.generate_uid()
        print(f"Votre ID est : {self.local_uid}")
        return self.local_pseudo, self.local_uid

    def start_server(self):
        try:
            self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.sock.bind((self.host, self.port))
            self.sock.listen(5)
            print(f"{self.local_pseudo} [{self.local_uid}] est maintenant le noeud.")
            print("En attente de connexions...")
            
            while self.running:
                try:
                    conn, addr = self.sock.accept()
                    
                    thread = threading.Thread(target=self.handle_client, args=(conn, addr), daemon=True)
                    thread.start()
                except Exception as e:
                    if self.running:
                        print(f"Erreur acceptation: {e}")
        except Exception as e:
            if self.running:
                print(f"Erreur serveur principal: {e}")
        finally:
            if not self.running:
                if self.sock:
                    try: self.sock.close()
                    except: pass

    def handle_client(self, conn, addr):
        conn.settimeout(1.0) 
        try:
            data = conn.recv(1024).decode()
            if not data:
                return
            
            parts = data.strip().split(':')
            if len(parts) >= 3:
                pseudo = parts[0]
                uid = parts[1]
                msg_type = parts[2]
                
                if msg_type == "JOIN":
                    user = User(conn, addr, pseudo, uid)
                    with self.lock:
                        self.users[addr] = user
                        self.current_node_users += 1
                        print(f"-> [{pseudo}#{uid}] a rejoint le noeud.")
                        print(f"Il y a {self.current_node_users} utilisateurs connectes")
                    
                    
                    self.broadcast(f"MSG:{pseudo}#{uid} a rejoint le noeud.", pseudo)
                    
                    
                    while self.running and addr in self.users:
                        try:
                            recvd = conn.recv(1024).decode()
                            if not recvd:
                                break
                            
                            parts_recv = recvd.strip().split(':', 2)
                            if len(parts_recv) >= 3:
                                p, u, msg = parts_recv
                                if p == self.local_pseudo and u == self.local_uid:
                                    
                                    pass 
                                print(f"[{p}#{u}]: {msg}")
                                self.broadcast(recvd.strip(), p)
                            elif recvd.strip().startswith("LEAVE"):
                                
                                pass
                        except socket.timeout:
                            continue
                        except Exception:
                            break

                elif msg_type == "MSG":
                    msg_content = ':'.join(parts[3:])
                    print(f"[{pseudo}#{uid}]: {msg_content}")
                    self.broadcast(recvd.strip(), pseudo) 
                    
                elif msg_type == "LEAVE":
                    self.remove_user(addr)
                    conn.close()

        except Exception as e:
            pass
        finally:
            self.remove_user(addr)
            try: conn.close()
            except: pass

    def remove_user(self, addr):
        with self.lock:
            if addr in self.users:
                user = self.users.pop(addr)
                self.current_node_users -= 1
                print(f"-> {user.pseudo} a quitte le noeud.")
                print(f"Il y a {self.current_node_users} utilisateurs connectes")
                self.broadcast(f"MSG:{user.pseudo} a quitte le noeud.", None)

    def broadcast(self, message, sender_pseudo):
        with self.lock:
            for addr, user in list(self.users.items()):
                try:
                    user.conn.sendall((message + "\n").encode())
                except:
                    pass

    def connect_to_node(self, host, port):
        try:
            self.client_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.client_sock.connect((host, port))
            
            
            join_msg = f"{self.local_pseudo}:{self.local_uid}:JOIN"
            self.client_sock.sendall(join_msg.encode())
            
            print("Vous etes tout seul, merci d'attendre...")
            
            
            def read_server():
                while self.running:
                    try:
                        data = self.client_sock.recv(1024).decode()
                        if not data: break
                        print(data)
                    except:
                        break
                
                if self.client_sock:
                    try: self.client_sock.close()
                    except: pass
                self.is_chatting = False

            t_read = threading.Thread(target=read_server, daemon=True)
            t_read.start()
            
            self.is_chatting = True
            print("--- Mode Chat Actif (Tape 'QUIT' pour sortir) ---")
            
            while self.is_chatting and self.running:
                try:
                    
                    cmd = input("> ").strip()
                    if cmd.upper() == "QUIT":
                        break
                    
                    if cmd:
                        msg_to_send = f"{self.local_pseudo}:{self.local_uid}:MSG:{cmd}"
                        self.client_sock.sendall(msg_to_send.encode())
                except EOFError:
                    break
                except Exception as e:
                    print(f"Erreur input: {e}")
                    break
            
            
            leave_msg = f"{self.local_pseudo}:{self.local_uid}:LEAVE"
            try:
                self.client_sock.sendall(leave_msg.encode())
            except:
                pass
            time.sleep(0.5) 
            self.is_chatting = False
            print("Deconnecte du noeud.")

        except Exception as e:
            print(f"Erreur de connexion : {e}")
            self.is_chatting = False

    def run(self):
        self.choose_identity()
        
        
        self.server_thread = threading.Thread(target=self.start_server, daemon=True)
        self.server_thread.start()
        
        while self.running:
            if not self.is_chatting:
                print("\n--- Menu pytopy ---")
                print("1. Creer/Rejoindre un noeud P2P")
                print("2. Quitter")
                
                choice = input("Choix : ").strip()
                
                if choice == "1":
                    mode = input("1 pour Creer, 2 pour Rejoindre : ").strip()
                    if mode == "1":
                        print("Le noeud est cree. Attendez les connexions.")
                        
                        time.sleep(1)
                    elif mode == "2":
                        host = input("Adresse du noeud (IP ou localhost) : ").strip()
                        port = 9999
                        try:
                            port_str = input("Port (defaut 9999) : ").strip()
                            if port_str:
                                port = int(port_str)
                        except:
                            pass
                        self.connect_to_node(host, port)
                    else:
                        print("Choix invalide.")
                elif choice == "2":
                    self.running = False
                    print("Fermeture de pytopy...")
                else:
                    print("Choix invalide.")
            else:
                
                time.sleep(0.1)

if __name__ == "__main__":
    node = P2PNode()
    try:
        node.run()
    except KeyboardInterrupt:
        node.running = False
        print("\nArret de pytopy.")