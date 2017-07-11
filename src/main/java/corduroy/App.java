package corduroy;

/**
 * App is a wrapper that can be used to launch nodes.
 */
public class App 
{
    /**
     * Runs the application.
     * @param args Arguments.
     */
    public static void main( String[] args )
    {
        Node node = new Node("127.0.0.1:8080");
        System.out.println(node.getHash(1));
    }
}
