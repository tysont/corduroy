package corduroy;

import lombok.Getter;
import lombok.SneakyThrows;

import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketException;
import java.util.HashSet;
import java.util.Hashtable;
import java.util.Map;
import java.util.Set;

/**
 * Node represents a single running instance where data can be stored and retrieved.
 */
public class Node implements Runnable {

    /**
     * Creates a node by passing in an {@link InetAddress} and an {@link int} that represents a Port.
     * @param port The port that the node will listen on.
     */
    @SneakyThrows
    public Node(int port) {
        this.inetAddress = InetAddress.getLocalHost();
        this.port = port;

        localTable = new Hashtable<String, Object>();
        requestHandlerThreads = new HashSet<Thread>();
        fingerTable = new Hashtable<Integer, String>();
    }

    /**
     * Initialize the node by building its finger table.
     * @param address The address of another node in the ring, to seed initialization.
     */
    public void initialize(String address) {
        fingerTable.put(1, address);
    }

    /**
     * Runs a node by calling listen so the node can be kicked off in a different thread.
     */
    public void run() {
        listen();
    }

    /**
     * Starts listening for messages on the specified port.
     */
    @SneakyThrows
    public void listen() {

        listenSocket = new ServerSocket(port);
        listening = true;

        while (listening) {

            try {

                System.out.println("Listening...");
                Socket clientSocket = listenSocket.accept();
                System.out.println("Socket established...");

                ObjectInputStream requestStream = new ObjectInputStream(clientSocket.getInputStream());
                Message requestMessage = (Message) requestStream.readObject();

                ObjectOutputStream responseStream = new ObjectOutputStream(clientSocket.getOutputStream());
                Handler requestHandler = new Handler(requestMessage, responseStream, this);

                Thread t = new Thread(requestHandler);
                requestHandlerThreads.add(t);
                t.start();

            }
            catch (SocketException ex) {
                System.out.println(ex.getMessage());
            }

        }
    }

    /**
     * Kills the node and stops it from listening.
     */
    @SneakyThrows
    public void kill() {

        listenSocket.close();
        listening = false;
    }

    /**
     * Sends a requestMessage to another node with the given address.
     * @param requestMessage The outbound requestMessage to send.
     * @param address The address to send the requestMessage to.
     * @return The inbound requestMessage that is received in response.
     */
    @SneakyThrows
    public static Message send(Message requestMessage, String address) {

        InetAddress host = Utility.getInetAddress(address);
        int port = Utility.getPort(address);

        Socket clientSocket = new Socket(host, port);
        ObjectOutputStream requestStream = new ObjectOutputStream(clientSocket.getOutputStream());
        requestStream.writeObject(requestMessage);

        ObjectInputStream responseStream = new ObjectInputStream(clientSocket.getInputStream());
        Message responseMessage = (Message)responseStream.readObject();
        clientSocket.close();
        return responseMessage;
    }

    /**
     * Gets the nodes address in the format 'host:port'.
     * @return The address.
     */
    public String getAddress() {
        return Utility.getAddress(inetAddress, port);
    }

    /**
     * Gets the hash value of the node with respect ot a specific ring number.
     * @param ring The number of the ring.
     * @return The hash value as a positive integer.
     */
    public int getHash(int ring) {
        try {
            return Utility.hash(getAddress(), ring);
        }
        catch (Exception ex) {
            return -1;
        }
    }

    /**
     * The {@link InetAddress} where the node is listening
     */
    private InetAddress inetAddress;

    /**
     * The {@link int} number of the Port where the node is listening.
     */
    private int port;

    /**
     * The {@link ServerSocket} where the node listens for new incoming connections.
     */
    private ServerSocket listenSocket;

    /**
     * Whether the node is listening.
     */
    private boolean listening;

    /**
     * A finger table of nodes to forward the request to.
     */
    @Getter
    private Map<Integer, String> fingerTable;

    /**
     * The local table of key/value pairs.
     */
    @Getter
    private Map<String, Object> localTable;

    /**
     * A list of active request handler threads.
     */
    @Getter
    private Set<Thread> requestHandlerThreads;
}
