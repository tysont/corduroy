package corduroy;

import lombok.SneakyThrows;

import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketException;

/**
 * Node represents a single running instance where data can be stored and retrieved.
 */
public class Node implements Runnable {

    /**
     * Creates a node by passing in an {@link InetAddress} and an {@link int} that represents a Port.
     * @param port
     */
    @SneakyThrows
    public Node(int port) {
        this.inetAddress = InetAddress.getLocalHost();
        this.port = port;
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

                ObjectOutputStream outToClient = new ObjectOutputStream(clientSocket.getOutputStream());
                ObjectInputStream inFromClient = new ObjectInputStream(clientSocket.getInputStream());

                Message inMessage = (Message) inFromClient.readObject();
                String text = (String)inMessage.getPayload();

                Message outMessage = new Message();
                outMessage.setPayload(text.toUpperCase());
                outToClient.writeObject(outMessage);

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
     * Sends a message to another node with the given address.
     * @param outMessage The outbound message to send.
     * @param address The address to send the message to.
     * @return The inbound message that is received in response.
     */
    @SneakyThrows
    public static Message send(Message outMessage, String address) {

        InetAddress host = Utility.getInetAddress(address);
        int port = Utility.getPort(address);

        Socket clientSocket = new Socket(host, port);
        ObjectOutputStream outToServer = new ObjectOutputStream(clientSocket.getOutputStream());
        ObjectInputStream inFromServer = new ObjectInputStream(clientSocket.getInputStream());

        outToServer.writeObject(outMessage);

        Message inMessage = (Message)inFromServer.readObject();
        clientSocket.close();
        return inMessage;
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
}
